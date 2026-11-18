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
