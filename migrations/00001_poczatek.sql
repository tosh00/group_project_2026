-- Schemat początkowy backendu (ADR 0008, ADR 0009).
-- Dane wejściowe, plan i raport walidatora trzymamy jako jsonb; uruchomienia mają zwykłe kolumny.

-- +goose Up
CREATE TABLE zestawy_danych (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    nazwa        text        NOT NULL,
    dane         jsonb       NOT NULL, -- treść dane_planu.json
    podsumowanie jsonb       NOT NULL, -- liczby tabel (API: Podsumowanie)
    ostrzezenia  jsonb       NOT NULL DEFAULT '[]', -- API: Ostrzezenie[]
    utworzono    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE uruchomienia (
    id              uuid             PRIMARY KEY DEFAULT gen_random_uuid(),
    zestaw_id       uuid             NOT NULL REFERENCES zestawy_danych (id) ON DELETE CASCADE,
    status          text             NOT NULL DEFAULT 'w_kolejce'
        CHECK (status IN ('w_kolejce', 'trwa', 'poprawny', 'z_naruszeniami', 'blad')),
    wagi            jsonb            NOT NULL,
    seed            bigint           NOT NULL,
    iteracje        integer          NOT NULL CHECK (iteracje >= 0),
    zatrzymane      boolean          NOT NULL DEFAULT false,
    liczba_naruszen integer,
    koszt           double precision,
    blad            text,
    utworzono       timestamptz      NOT NULL DEFAULT now(),
    rozpoczeto      timestamptz,
    zakonczono      timestamptz
);

CREATE INDEX uruchomienia_zestaw_idx ON uruchomienia (zestaw_id, utworzono DESC);
CREATE INDEX uruchomienia_niedokonczone_idx ON uruchomienia (status) WHERE status IN ('w_kolejce', 'trwa');

CREATE TABLE plany (
    uruchomienie_id uuid        PRIMARY KEY REFERENCES uruchomienia (id) ON DELETE CASCADE,
    plan            jsonb       NOT NULL, -- plan.json v1
    raport          jsonb       NOT NULL, -- raport walidatora
    utworzono       timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE plany;
DROP TABLE uruchomienia;
DROP TABLE zestawy_danych;
