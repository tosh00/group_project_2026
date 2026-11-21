-- name: DodajZestaw :one
INSERT INTO zestawy_danych (nazwa, dane, podsumowanie, ostrzezenia)
VALUES ($1, $2, $3, $4)
RETURNING id, nazwa, podsumowanie, ostrzezenia, utworzono;

-- name: ListaZestawow :many
SELECT id, nazwa, podsumowanie, ostrzezenia, utworzono
FROM zestawy_danych
ORDER BY utworzono DESC, id;

-- name: PobierzZestaw :one
SELECT id, nazwa, podsumowanie, ostrzezenia, utworzono
FROM zestawy_danych
WHERE id = $1;

-- name: PobierzDaneZestawu :one
SELECT dane FROM zestawy_danych WHERE id = $1;

-- name: UsunZestaw :execrows
-- Usuwa zestaw z uruchomieniami i planami (ON DELETE CASCADE), ale nie w trakcie generowania:
-- zestaw z uruchomieniem w kolejce albo trwającym zostaje (0 usuniętych wierszy).
DELETE FROM zestawy_danych z
WHERE z.id = $1
  AND NOT EXISTS (
    SELECT 1 FROM uruchomienia u
    WHERE u.zestaw_id = z.id AND u.status IN ('w_kolejce', 'trwa')
  );
