// v1_deadlock.go — V1: versão concorrente "ingênua", sujeita a deadlock.
// Cada garfo é um sync.Mutex. Todo filósofo pega primeiro o garfo da ESQUERDA
// e depois o da DIREITA. Se os cinco pegarem o esquerdo ao mesmo tempo, cada um
// fica esperando o garfo que está com o vizinho: espera circular -> deadlock.
package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Estado observável de cada filósofo, para mostrar a evidência do deadlock.
type estadoV1 struct {
	mu        sync.Mutex
	segura    [N][]int // garfos que cada filósofo possui
	esperando [N]int   // garfo pelo qual está esperando (-1 = nenhum)
}

func (e *estadoV1) set(id int, segura []int, esperando int) {
	e.mu.Lock()
	e.segura[id] = segura
	e.esperando[id] = esperando
	e.mu.Unlock()
}

func V1(pausa time.Duration, vigia bool) Resultado {
	m := &Metricas{}
	func() {
		defer trackTime(time.Now(), "V1", m) // mede o tempo total do jantar
		garfos := make([]sync.Mutex, N)
		est := &estadoV1{}
		for i := range est.esperando {
			est.esperando[i] = -1
		}
		var progresso int64 // incrementa a cada refeição; parado = suspeita de deadlock

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
					est.set(id, nil, esq)
					garfos[esq].Lock()  // pega o garfo esquerdo...
					marca(id, "g", esq) // g<k> = pegou o garfo k (o esquerdo)
					est.set(id, []int{esq}, dir)
					if pausa > 0 {
						time.Sleep(pausa) // ...janela que torna o deadlock reproduzível
					}
					garfos[dir].Lock() // ...e SEGURA o esquerdo enquanto espera o direito
					est.set(id, []int{esq, dir}, -1)
					m.RegistraEspera(id, time.Since(ini))

					m.ComecaComer()
					marca(id, "E", c+1)
					time.Sleep(TempoCome) // comer
					m.TerminaComer(id)
					atomic.AddInt64(&progresso, 1)

					garfos[dir].Unlock() // liberar recursos
					garfos[esq].Unlock()
					est.set(id, nil, -1)
				}
			}(id)
		}

		if vigia {
			// Vigia: se ninguém comer por 2 s, declara deadlock e imprime as evidências.
			go func() {
				ultimo := int64(-1)
				for {
					time.Sleep(2 * time.Second)
					p := atomic.LoadInt64(&progresso)
					if p == ultimo {
						imprimeDeadlock(est, m, p)
						os.Exit(2)
					}
					ultimo = p
				}
			}()
		}
		// Sem vigia, quem detecta o deadlock é o próprio runtime do Go:
		// "fatal error: all goroutines are asleep - deadlock!"
		wg.Wait()
	}()
	return m.Consolida("V1")
}

func imprimeDeadlock(est *estadoV1, m *Metricas, refeicoes int64) {
	est.mu.Lock()
	defer est.mu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Println("==================== DEADLOCK DETECTADO ====================")
	fmt.Printf("Nenhum filósofo comeu nos últimos 2 s (refeições até aqui: %d)\n\n", refeicoes)
	// dono[g] = filósofo que segura o garfo g (-1 = livre)
	dono := [N]int{-1, -1, -1, -1, -1}
	for i := 0; i < N; i++ {
		for _, g := range est.segura[i] {
			dono[g] = i
		}
	}
	fmt.Println("Filósofo | Garfos que possui | Esperando pelo garfo | Ciclos")
	for i := 0; i < N; i++ {
		esp := "-"
		if g := est.esperando[i]; g >= 0 {
			esp = fmt.Sprintf("%d (com F%d)", g, dono[g])
		}
		fmt.Printf("   F%d    | %-17s | %-20s | %d\n", i, fmt.Sprint(est.segura[i]), esp, m.ciclos[i])
	}
	// Segue a relação "Fi espera garfo que está com Fj" a partir de F0.
	fmt.Print("\nCadeia de espera: F0")
	atual := 0
	for passo := 0; passo < N; passo++ {
		g := est.esperando[atual]
		if g < 0 || dono[g] < 0 {
			break
		}
		atual = dono[g]
		fmt.Printf(" -> F%d", atual)
		if atual == 0 {
			fmt.Print("  (ciclo fechado: espera circular)")
			break
		}
	}
	fmt.Println()
	fmt.Println("============================================================")
}
