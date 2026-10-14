// v3_ordenacao.go — V3: sem deadlock e com mais concorrência.
// Estratégia: ordenação global dos recursos (hierarquia de Dijkstra).
// Cada filósofo pega SEMPRE primeiro o garfo de menor número e depois o de maior.
// Para F0..F3 isso é "esquerdo, depois direito"; para F4 (garfos 4 e 0) a ordem se
// inverte: ele pega o 0 antes do 4. Isso quebra a espera circular, a única
// condição de Coffman que a V1 tinha e que é fácil de remover, e só serializa
// quem realmente disputa o mesmo garfo (vizinhos).
//
// Variante V3b (par/ímpar, sugerida no artigo da ZeroBone): filósofos pares pegam
// esquerdo->direito e ímpares direito->esquerdo. Vizinhos passam a disputar
// primeiro o MESMO garfo, o que também impede a espera circular e encurta as
// cadeias de espera.
package main

import (
	"sync"
	"time"
)

func V3(parImpar bool) Resultado {
	nome := "V3"
	if parImpar {
		nome = "V3b"
	}
	m := &Metricas{}
	func() {
		defer trackTime(time.Now(), nome, m) // mede o tempo total do jantar
		garfos := make([]sync.Mutex, N)

		var wg sync.WaitGroup
		for id := 0; id < N; id++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				primeiro, segundo := garfoEsq(id), garfoDir(id)
				if parImpar {
					if id%2 == 1 { // ímpar: direito primeiro
						primeiro, segundo = segundo, primeiro
					}
				} else if primeiro > segundo { // hierarquia: só acontece com F4
					primeiro, segundo = segundo, primeiro
				}
				for c := 0; c < CICLOS; c++ {
					marca(id, "T", c+1)
					time.Sleep(TempoPensa) // pensar

					ini := time.Now()
					garfos[primeiro].Lock() // menor número primeiro
					garfos[segundo].Lock()  // depois o maior
					m.RegistraEspera(id, time.Since(ini))

					m.ComecaComer()
					marca(id, "E", c+1)
					time.Sleep(TempoCome) // comer
					m.TerminaComer(id)

					garfos[segundo].Unlock() // liberar recursos
					garfos[primeiro].Unlock()
				}
			}(id)
		}
		wg.Wait()
	}()
	return m.Consolida(nome)
}
