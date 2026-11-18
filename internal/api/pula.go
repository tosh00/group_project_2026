package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/solver"
	"github.com/tosh00/group_project_2026/internal/store"
)

// ErrKolejkaPelna: pula nie przyjmie więcej uruchomień, dopóki któreś się nie zacznie.
var ErrKolejkaPelna = errors.New("kolejka uruchomień jest pełna")

// Typy zdarzeń SSE (GET /api/uruchomienia/{id}/postep).
const (
	ZdarzenieStan   = "stan"
	ZdarzeniePostep = "postep"
	ZdarzenieKoniec = "koniec"
)

// Zdarzenie to jedno zdarzenie SSE: Dane to Uruchomienie albo Postep.
type Zdarzenie struct {
	Typ  string
	Dane any
}

// UstawieniaPuli konfigurują pulę workerów.
type UstawieniaPuli struct {
	Workery        int           // równoległe uruchomienia (każde zajmuje WorkerySolvera rdzeni); domyślnie 1
	Kolejka        int           // ile uruchomień może czekać; domyślnie 16
	WorkerySolvera int           // solver.ZWorkerami; domyślnie solver.DomyslniWorkerzy
	CoIlePostep    time.Duration // najczęstsze zdarzenie "postep"; domyślnie 250 ms
}

// Pula uruchamia solver w tle dla uruchomień z kolejki, zapisuje wyniki i rozsyła postęp.
type Pula struct {
	repo      Repozytorium
	walidator Walidator
	log       *slog.Logger
	ust       UstawieniaPuli

	kolejka chan uuid.UUID
	ctx     context.Context
	stop    context.CancelFunc
	wg      sync.WaitGroup

	mu    sync.Mutex
	biegi map[uuid.UUID]*bieg // uruchomienia w kolejce i trwające
}

// bieg to stan uruchomienia w pamięci, od zgłoszenia do zakończenia. Pola chroni Pula.mu.
type bieg struct {
	u          store.Uruchomienie // ostatni znany wiersz (do zdarzeń "stan")
	faza       string
	anuluj     context.CancelFunc // nil, dopóki czeka w kolejce
	zatrzymane bool               // użytkownik poprosił o zatrzymanie
	sub        map[chan Zdarzenie]struct{}
}

// NowaPula tworzy pulę i uruchamia jej workery. Zamknij kończy pracę.
func NowaPula(repo Repozytorium, w Walidator, log *slog.Logger, ust UstawieniaPuli) *Pula {
	if ust.Workery <= 0 {
		ust.Workery = 1
	}
	if ust.Kolejka <= 0 {
		ust.Kolejka = 16
	}
	if ust.WorkerySolvera <= 0 {
		ust.WorkerySolvera = solver.DomyslniWorkerzy
	}
	if ust.CoIlePostep <= 0 {
		ust.CoIlePostep = 250 * time.Millisecond
	}
	ctx, stop := context.WithCancel(context.Background())
	p := &Pula{
		repo: repo, walidator: w, log: log, ust: ust,
		kolejka: make(chan uuid.UUID, ust.Kolejka),
		ctx:     ctx, stop: stop,
		biegi: make(map[uuid.UUID]*bieg),
	}
	for range ust.Workery {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

// Zamknij przerywa trwające uruchomienia (zapisują najlepszy plan, jeśli konstrukcja się skończyła)
// i czeka na workery do końca ctx. Uruchomienia z kolejki zostają w bazie jako w_kolejce;
// przy następnym starcie serwer oznacza je jako błąd (store.PrzerwijNiedokonczone).
func (p *Pula) Zamknij(ctx context.Context) error {
	p.stop()
	gotowe := make(chan struct{})
	go func() { p.wg.Wait(); close(gotowe) }()
	select {
	case <-gotowe:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Zglos dodaje uruchomienie (status w_kolejce w bazie) do kolejki.
func (p *Pula) Zglos(u store.Uruchomienie) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case p.kolejka <- u.ID:
		p.biegi[u.ID] = &bieg{u: u, sub: make(map[chan Zdarzenie]struct{})}
		return nil
	default:
		return ErrKolejkaPelna
	}
}

// Faza zwraca fazę trwającego uruchomienia ("" poza pulą).
func (p *Pula) Faza(id uuid.UUID) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if b := p.biegi[id]; b != nil {
		return b.faza
	}
	return ""
}

// Zatrzymaj przerywa uruchomienie. Czekające kończy od razu błędem, trwającemu anuluje kontekst
// (solver odda najlepszy plan). false: uruchomienia nie ma w puli (już się skończyło).
func (p *Pula) Zatrzymaj(ctx context.Context, id uuid.UUID) (bool, error) {
	p.mu.Lock()
	b := p.biegi[id]
	if b == nil {
		p.mu.Unlock()
		return false, nil
	}
	if b.zatrzymane {
		p.mu.Unlock()
		return true, nil
	}
	b.zatrzymane = true
	if b.anuluj != nil {
		b.anuluj()
		p.mu.Unlock()
		return true, nil
	}
	p.mu.Unlock()

	// Czekało w kolejce: worker je pominie (zatrzymane = true), stan końcowy zapisujemy tutaj.
	opis := "zatrzymane przed rozpoczęciem"
	u, err := p.repo.Zakoncz(ctx, store.Zakonczenie{ID: id, Status: store.StatusBlad, Zatrzymane: true, Blad: &opis})
	if err != nil && !errors.Is(err, store.ErrZlyStan) {
		return true, err
	}
	if err != nil {
		u, err = p.repo.Uruchomienie(ctx, id)
		if err != nil {
			return true, err
		}
	}
	p.zakoncz(id, u)
	return true, nil
}

// Subskrybuj zwraca kanał zdarzeń uruchomienia. ok = false: uruchomienia nie ma w puli.
// Kanał zamyka się po zdarzeniu "koniec" albo gdy odbiorca nie nadąża; wtedy trzeba wywołać przestan.
func (p *Pula) Subskrybuj(id uuid.UUID) (zdarzenia <-chan Zdarzenie, przestan func(), ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	b := p.biegi[id]
	if b == nil {
		return nil, nil, false
	}
	ch := make(chan Zdarzenie, 32)
	b.sub[ch] = struct{}{}
	ch <- Zdarzenie{Typ: ZdarzenieStan, Dane: uruchomienieDTO(b.u, b.faza)}
	przestan = func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if _, jest := b.sub[ch]; jest {
			delete(b.sub, ch)
			close(ch)
		}
	}
	return ch, przestan, true
}

// wyslij rozsyła zdarzenie do subskrybentów biegu. Wołane z p.mu.
// Zdarzenia "postep" można zgubić; odbiorcę, który nie nadąża ze "stan", odłączamy (EventSource połączy się ponownie).
func (b *bieg) wyslij(z Zdarzenie) {
	for ch := range b.sub {
		select {
		case ch <- z:
		default:
			if z.Typ != ZdarzeniePostep {
				delete(b.sub, ch)
				close(ch)
			}
		}
	}
}

func (p *Pula) zmienStan(id uuid.UUID, zmien func(b *bieg)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if b := p.biegi[id]; b != nil {
		zmien(b)
		b.wyslij(Zdarzenie{Typ: ZdarzenieStan, Dane: uruchomienieDTO(b.u, b.faza)})
	}
}

// zakoncz wysyła "koniec", zamyka subskrypcje i usuwa bieg z puli.
func (p *Pula) zakoncz(id uuid.UUID, u store.Uruchomienie) {
	p.mu.Lock()
	defer p.mu.Unlock()
	b := p.biegi[id]
	if b == nil {
		return
	}
	b.wyslij(Zdarzenie{Typ: ZdarzenieKoniec, Dane: uruchomienieDTO(u, "")})
	for ch := range b.sub {
		close(ch)
	}
	delete(p.biegi, id)
}

func (p *Pula) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case id := <-p.kolejka:
			p.uruchom(id)
		}
	}
}

// uruchom wykonuje jedno uruchomienie od w_kolejce do stanu końcowego.
func (p *Pula) uruchom(id uuid.UUID) {
	p.mu.Lock()
	b := p.biegi[id]
	if b == nil || b.zatrzymane {
		p.mu.Unlock()
		return
	}
	ctx, anuluj := context.WithCancel(p.ctx)
	defer anuluj()
	b.anuluj = anuluj
	b.faza = FazaKonstrukcja
	p.mu.Unlock()

	// Zapis do bazy nie może zależeć od anulowania uruchomienia.
	bazaCtx := context.WithoutCancel(ctx)
	log := p.log.With("uruchomienie", id)

	u, err := p.repo.Rozpocznij(bazaCtx, id)
	if err != nil {
		log.Error("rozpoczęcie uruchomienia", "blad", err)
		if u, err := p.repo.Uruchomienie(bazaCtx, id); err == nil {
			p.zakoncz(id, u)
		} else {
			p.zakoncz(id, b.u)
		}
		return
	}
	p.zmienStan(id, func(b *bieg) { b.u = u })
	log.Info("start", "seed", u.Seed, "iteracje", u.Iteracje)

	z := p.wykonaj(ctx, id, u)
	koniec, err := p.repo.Zakoncz(bazaCtx, z)
	if err != nil {
		log.Error("zapis wyniku", "blad", err)
		if koniec, err = p.repo.Uruchomienie(bazaCtx, id); err != nil {
			koniec = u
		}
	}
	log.Info("koniec", "status", koniec.Status, "zatrzymane", koniec.Zatrzymane)
	p.zakoncz(id, koniec)
}

// wykonaj uruchamia solver i walidator. Zwraca stan końcowy do zapisania (bez paniki, błędy jako status blad).
func (p *Pula) wykonaj(ctx context.Context, id uuid.UUID, u store.Uruchomienie) (z store.Zakonczenie) {
	z.ID = id
	blad := func(format string, args ...any) store.Zakonczenie {
		opis := fmt.Sprintf(format, args...)
		z.Status, z.Blad = store.StatusBlad, &opis
		return z
	}
	czyZatrzymane := func() bool {
		p.mu.Lock()
		defer p.mu.Unlock()
		b := p.biegi[id]
		return (b != nil && b.zatrzymane) || p.ctx.Err() != nil
	}
	defer func() {
		if r := recover(); r != nil {
			p.log.Error("panika w uruchomieniu", "uruchomienie", id, "panika", r)
			z = blad("błąd wewnętrzny solvera: %v", r)
		}
	}()

	surowe, err := p.repo.DaneZestawu(context.WithoutCancel(ctx), u.ZestawID)
	if err != nil {
		return blad("odczyt danych zestawu: %v", err)
	}
	dane, err := model.Dekoduj(bytes.NewReader(surowe))
	if err != nil {
		return blad("dane zestawu: %v", err)
	}
	wagi := solver.DomyslneWagi()
	if err := json.Unmarshal(u.Wagi, &wagi); err != nil {
		return blad("wagi: %v", err)
	}

	start := time.Now()
	ag := &agregator{
		pula: p, id: id, start: start, coIle: p.ust.CoIlePostep,
		ostatnie: make([]solver.Postep, p.ust.WorkerySolvera),
		iteracji: int(u.Iteracje) * p.ust.WorkerySolvera,
	}
	plan, err := solver.Solve(ctx, dane, wagi,
		solver.ZSeed(u.Seed),
		solver.ZIteracjami(int(u.Iteracje)),
		solver.ZWorkerami(p.ust.WorkerySolvera),
		solver.ZPostepem(ag.przyjmij),
	)
	z.Zatrzymane = czyZatrzymane()
	if err != nil {
		if z.Zatrzymane {
			return blad("zatrzymane przed końcem konstrukcji")
		}
		return blad("przerwane przed końcem konstrukcji: %v", err)
	}
	ag.wyslijKoncowy()

	p.zmienStan(id, func(b *bieg) { b.faza = FazaWalidacja })
	raport := p.ocen(dane, &plan, wagi)

	var bufor bytes.Buffer
	if err := model.ZapiszPlan(&bufor, &plan); err != nil {
		return blad("zapis planu: %v", err)
	}
	raportJSON, err := json.Marshal(raport)
	if err != nil {
		return blad("zapis raportu: %v", err)
	}
	liczba := int32(raport.LiczbaNaruszen)
	koszt := plan.Meta.Koszt
	z.Status = raport.Status
	z.LiczbaNaruszen, z.Koszt = &liczba, &koszt
	z.Plan, z.Raport = bufor.Bytes(), raportJSON
	return z
}

// ocen buduje raport walidatora: naruszenia twarde i rozbicie kosztu miękkiego.
func (p *Pula) ocen(dane *model.Dane, plan *model.Plan, wagi solver.Wagi) Raport {
	naruszenia := p.walidator.Sprawdz(dane, plan)
	if naruszenia == nil {
		naruszenia = []Naruszenie{}
	}
	for i := range naruszenia {
		if naruszenia[i].Nazwa == "" {
			naruszenia[i].Nazwa = NazwaReguly(naruszenia[i].Regula)
		}
	}
	kryteria, err := solver.RozbicieKosztu(dane, *plan, wagi)
	if err != nil {
		p.log.Warn("rozbicie kosztu", "blad", err)
		kryteria = map[string]float64{}
	}
	r := Raport{
		Status:         store.StatusPoprawny,
		LiczbaNaruszen: len(naruszenia),
		Naruszenia:     naruszenia,
		KosztMiekki:    plan.Meta.Koszt,
		KosztKryteria:  kryteria,
		Walidator:      p.walidator.Nazwa(),
	}
	if len(naruszenia) > 0 {
		r.Status = store.StatusZNaruszeniami
	}
	return r
}

// agregator łączy postęp workerów solvera w jedno zdarzenie "postep", najwyżej co coIle.
type agregator struct {
	pula     *Pula
	id       uuid.UUID
	start    time.Time
	coIle    time.Duration
	iteracji int

	mu       sync.Mutex
	ostatnie []solver.Postep // ostatni postęp każdego workera solvera
	wyslano  time.Time
	byl      bool
}

func (a *agregator) przyjmij(pp solver.Postep) {
	a.mu.Lock()
	if pp.Pracownik >= 0 && pp.Pracownik < len(a.ostatnie) {
		a.ostatnie[pp.Pracownik] = pp
	}
	pierwszy := !a.byl
	a.byl = true
	teraz := time.Now()
	if !pierwszy && teraz.Sub(a.wyslano) < a.coIle {
		a.mu.Unlock()
		return
	}
	a.wyslano = teraz
	zd := a.postep(teraz)
	a.mu.Unlock()

	a.pula.mu.Lock()
	defer a.pula.mu.Unlock()
	b := a.pula.biegi[a.id]
	if b == nil {
		return
	}
	if pierwszy {
		b.faza = FazaOptymalizacja
		b.wyslij(Zdarzenie{Typ: ZdarzenieStan, Dane: uruchomienieDTO(b.u, b.faza)})
	}
	b.wyslij(Zdarzenie{Typ: ZdarzeniePostep, Dane: zd})
}

// wyslijKoncowy wysyła ostatni stan postępu po zakończeniu solvera (jeśli była faza optymalizacji).
func (a *agregator) wyslijKoncowy() {
	a.mu.Lock()
	if !a.byl {
		a.mu.Unlock()
		return
	}
	zd := a.postep(time.Now())
	a.mu.Unlock()
	a.pula.mu.Lock()
	defer a.pula.mu.Unlock()
	if b := a.pula.biegi[a.id]; b != nil {
		b.wyslij(Zdarzenie{Typ: ZdarzeniePostep, Dane: zd})
	}
}

// postep: wołane z a.mu.
func (a *agregator) postep(teraz time.Time) Postep {
	wynik := Postep{Faza: FazaOptymalizacja, Iteracji: a.iteracji, CzasMs: teraz.Sub(a.start).Milliseconds()}
	najlepszy := -1
	for i, pp := range a.ostatnie {
		wynik.Iteracja += pp.Iteracja
		if pp.Iteracji == 0 { // worker jeszcze nic nie zgłosił
			continue
		}
		if najlepszy < 0 || lepszy(pp, a.ostatnie[najlepszy]) {
			najlepszy = i
		}
	}
	if najlepszy >= 0 {
		wynik.Koszt = a.ostatnie[najlepszy].Koszt
		wynik.Naruszenia = a.ostatnie[najlepszy].Naruszenia
	}
	return wynik
}

// lepszy: najpierw mniej naruszeń twardych, potem niższy koszt (jak w solverze, ADR 0006).
func lepszy(a, b solver.Postep) bool {
	if a.Naruszenia != b.Naruszenia {
		return a.Naruszenia < b.Naruszenia
	}
	return a.Koszt < b.Koszt
}
