# 0001. Repozytorium jako wspólna pamięć zespołu i agentów

- Status: przyjęta
- Data: 2026-10-24
- Decydują: zespół

## Kontekst

Każda osoba pracuje osobno, z własnym agentem (Claude Code). Pamięć agenta (`~/.claude`) jest
lokalna, więc ustalenia jednej osoby nie docierają do agentów pozostałych.

## Decyzja

Wspólna wiedza żyje w repozytorium:

- `CLAUDE.md`: zasady, które każdy agent czyta na starcie,
- `docs/decisions/`: decyzje (ADR),
- `docs/contracts/` i pliki kontraktów: interfejsy między modułami,
- `docs/notes.md`: wiedza o danych i problemie,
- GitHub Issues: zadania i status.

Agenty nie komunikują się bezpośrednio. Ustalenia przechodzą przez PR-y i issues.

## Rozważane opcje

- Wspólna baza pamięci lub kanał między agentami: trudne do przeglądania, brak review, łatwo o chaos.
- Repozytorium: wersjonowane, z review, każdy agent ma do niego dostęp.

## Konsekwencje

Każde ustalenie, które ma przetrwać sesję, trafia do repo w PR. Kontrakty powstają wcześnie (etap 0),
żeby moduły dało się rozwijać równolegle.
