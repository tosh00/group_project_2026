-- name: DodajUruchomienie :one
INSERT INTO uruchomienia (zestaw_id, wagi, seed, iteracje)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: PobierzUruchomienie :one
SELECT * FROM uruchomienia WHERE id = $1;

-- name: ListaUruchomien :many
SELECT * FROM uruchomienia
WHERE sqlc.narg(zestaw_id)::uuid IS NULL OR zestaw_id = sqlc.narg(zestaw_id)::uuid
ORDER BY utworzono DESC, id;

-- name: RozpocznijUruchomienie :one
-- Przejście w_kolejce → trwa. Brak wiersza oznacza, że uruchomienie już nie czeka (np. zatrzymane).
UPDATE uruchomienia
SET status = 'trwa', rozpoczeto = now()
WHERE id = $1 AND status = 'w_kolejce'
RETURNING *;

-- name: ZakonczUruchomienie :one
-- Stan końcowy; tylko z w_kolejce albo trwa, żeby nie nadpisać już zakończonego.
UPDATE uruchomienia
SET status          = $2,
    zatrzymane      = $3,
    liczba_naruszen = $4,
    koszt           = $5,
    blad            = $6,
    zakonczono      = now()
WHERE id = $1 AND status IN ('w_kolejce', 'trwa')
RETURNING *;

-- name: PrzerwijNiedokonczone :execrows
-- Przy starcie serwera: uruchomienia, których nikt już nie dokończy.
UPDATE uruchomienia
SET status = 'blad', blad = $1, zakonczono = now()
WHERE status IN ('w_kolejce', 'trwa');

-- name: ZapiszPlan :exec
INSERT INTO plany (uruchomienie_id, plan, raport) VALUES ($1, $2, $3);

-- name: PobierzPlan :one
SELECT plan FROM plany WHERE uruchomienie_id = $1;

-- name: PobierzRaport :one
SELECT raport FROM plany WHERE uruchomienie_id = $1;
