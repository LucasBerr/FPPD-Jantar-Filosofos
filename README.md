# T1 — Jantar dos Filósofos em Go (FPPD / PUCRS)

5 filósofos, 20 ciclos cada (pensar → pegar garfos → comer → soltar), pensar = 20 ms, comer = 30 ms
(mesma proporção 2:3 da base `filo_seq.go` do Prof. De Rose), iguais em todas as versões.

| Versão | Arquivo | Estratégia | Mecanismo |
|---|---|---|---|
| V0 | `v0_sequencial.go` | Sequencial (baseline, adaptada do Moodle) | nenhum |
| V1 | `v1_deadlock.go` | Esquerdo → direito, sujeita a deadlock | `sync.Mutex` por garfo |
| V2 | `v2_garcom_unico.go` | Trava global da mesa (sincronização excessiva) | `sync.Mutex` global + por garfo |
| V3 | `v3_ordenacao.go` | Ordenação hierárquica dos garfos (menor primeiro) | `sync.Mutex` por garfo |
| V3b | `v3_ordenacao.go` | Variante par/ímpar | `sync.Mutex` por garfo |

Métricas e tempo total (padrão `defer trackTime(time.Now(), ...)`) ficam em `comum.go`.

## Executar

```bash
go build -o filosofos .
./filosofos -v 0            # V0
./filosofos -v 1            # V1 (um vigia imprime a tabela do deadlock se travar)
./filosofos -v 1 -pausa 5ms # V1 com janela entre os garfos: deadlock reproduzível
./filosofos -v 1 -pausa 5ms -semvigia  # o runtime do Go acusa o deadlock
./filosofos -v 2            # V2
./filosofos -v 3            # V3
./filosofos -v 4            # V3b
# -log mostra o rastro em colunas: T<ciclo> pensando, E<ciclo> comendo, g<k> pegou garfo k
```

Experimento completo (5 execuções por versão + testes da V1): `bash run_experimentos.sh`.

## Referências
- Base sequencial `filo_seq.go` — Prof. Cesar De Rose (Moodle)
- Measuring execution time in Go — Coderwall
- The Dining Philosophers problem and different ways of solving it — ZeroBone
- The Dining Philosophers — Baeldung on Computer Science
