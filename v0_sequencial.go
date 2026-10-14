// v0_sequencial.go — V0: baseline sem concorrência.
// Adaptado de filo_seq.go (Prof. Cesar De Rose, Moodle): mesma estrutura
// Philosopher(id, iteration) chamada em laço "rodadas x filósofos", agora com
// 20 rodadas, tempos em ms e coleta das métricas do trabalho.
// Como é sequencial, o tempo esperado é rodadas*filósofos*(pensar+comer)
// = 20 * 5 * 50 ms = 5 s.
package main

import "time"

func PhilosopherSeq(id, iteration int, m *Metricas) {
	// think
	marca(id, "T", iteration)
	time.Sleep(TempoPensa)

	// adquirir recursos: imediato, pois ninguém mais executa ao mesmo tempo
	ini := time.Now()
	m.RegistraEspera(id, time.Since(ini))

	// eat
	m.ComecaComer()
	marca(id, "E", iteration)
	time.Sleep(TempoCome)
	m.TerminaComer(id) // liberar recursos
}

func V0() Resultado {
	m := &Metricas{}
	func() {
		defer trackTime(time.Now(), "V0", m) // mede o tempo total do jantar
		for i := 1; i <= CICLOS; i++ {       // rounds
			for j := 0; j < N; j++ { // philosophers
				PhilosopherSeq(j, i, m)
			}
		}
	}()
	return m.Consolida("V0")
}
