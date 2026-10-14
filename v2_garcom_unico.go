// v2_garcom_unico.go — V2: sem deadlock, mas com sincronização excessiva.
// Um único mutex global ("a mesa") protege TODO o trecho pegar garfos -> comer ->
// soltar garfos. Como só um filósofo por vez consegue passar pela mesa, a espera
// circular é impossível (ninguém segura um garfo esperando outro), mas também é
// impossível que dois filósofos NÃO vizinhos comam ao mesmo tempo.
package main

import (
	"sync"
	"time"
)

func V2() Resultado {
	m := &Metricas{}
	func() {
		defer trackTime(time.Now(), "V2", m) // mede o tempo total do jantar
		garfos := make([]sync.Mutex, N)
		var mesa sync.Mutex // trava global: a fonte da sincronização excessiva

		var wg sync.WaitGroup
		for id := 0; id < N; id++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				esq, dir := garfoEsq(id), garfoDir(id)
				for c := 0; c < CICLOS; c++ {
					marca(id, "T", c+1)
					time.Sleep(TempoPensa) // pensar

					ini := time.Now()
					mesa.Lock() // pede a mesa inteira só para si
					garfos[esq].Lock()
					garfos[dir].Lock()
					m.RegistraEspera(id, time.Since(ini))

					m.ComecaComer()
					marca(id, "E", c+1)
					time.Sleep(TempoCome) // comer — ainda segurando a mesa!
					m.TerminaComer(id)

					garfos[dir].Unlock()
					garfos[esq].Unlock()
					mesa.Unlock() // só agora outro filósofo pode tentar
				}
			}(id)
		}
		wg.Wait()
	}()
	return m.Consolida("V2")
}
