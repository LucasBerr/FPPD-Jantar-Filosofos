// comum.go — parâmetros do experimento e coleta de métricas, compartilhados por V0–V3.
package main

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Parâmetros idênticos para todas as versões (exigência do enunciado).
const (
	N      = 5  // número de filósofos (e de garfos)
	CICLOS = 20 // ciclos por filósofo
	// Mesma proporção da base do Prof. De Rose (pensar 2, comer 3), em ms em vez de s,
	// para que 20 ciclos x 5 filósofos caibam em poucos segundos.
	TempoPensa = 20 * time.Millisecond // duração simulada de "pensar"
	TempoCome  = 30 * time.Millisecond // duração simulada de "comer"
)

// marca imprime no formato de colunas da base filo_seq.go do professor:
// cada filósofo tem sua coluna (id*6 espaços) e escreve T<ciclo> ou E<ciclo>.
// Ligado com a flag -log; desligado nas medições para não interferir no tempo.
func marca(id int, evento string, ciclo int) {
	if !verbose {
		return
	}
	saida.Lock()
	fmt.Printf("%s %s%d\n", strings.Repeat(" ", id*6), evento, ciclo)
	saida.Unlock()
}

var saida sync.Mutex // evita que linhas de goroutines diferentes se misturem

// Garfo i fica entre o filósofo i (à sua esquerda) e o filósofo i+1.
func garfoEsq(i int) int { return i }
func garfoDir(i int) int { return (i + 1) % N }

// Metricas guarda tudo que o enunciado pede.
type Metricas struct {
	mu         sync.Mutex
	somaEspera [N]time.Duration // soma das esperas de cada filósofo
	maxEspera  [N]time.Duration // maior espera de cada filósofo
	ciclos     [N]int           // ciclos concluídos por filósofo

	comendoAgora int32 // quantos estão comendo neste instante
	maxComendo   int32 // pico de filósofos comendo ao mesmo tempo

	total time.Duration // preenchido por trackTime
}

// RegistraEspera é chamado logo depois que o filósofo obtém os dois garfos.
func (m *Metricas) RegistraEspera(id int, d time.Duration) {
	m.mu.Lock()
	m.somaEspera[id] += d
	if d > m.maxEspera[id] {
		m.maxEspera[id] = d
	}
	m.mu.Unlock()
}

// ComecaComer/TerminaComer mantêm o contador de simultaneidade (atômico).
func (m *Metricas) ComecaComer() {
	agora := atomic.AddInt32(&m.comendoAgora, 1)
	for {
		max := atomic.LoadInt32(&m.maxComendo)
		if agora <= max || atomic.CompareAndSwapInt32(&m.maxComendo, max, agora) {
			return
		}
	}
}

func (m *Metricas) TerminaComer(id int) {
	atomic.AddInt32(&m.comendoAgora, -1)
	m.mu.Lock()
	m.ciclos[id]++
	m.mu.Unlock()
}

// trackTime segue o padrão do artigo "Measuring execution time in Go" (Coderwall):
// é chamado com defer no início da função medida; time.Now() é avaliado na hora do
// defer, e time.Since roda quando a função retorna, por qualquer caminho.
func trackTime(inicio time.Time, nome string, m *Metricas) {
	m.total = time.Since(inicio)
	if verbose {
		fmt.Printf("%s took %s\n", nome, m.total)
	}
}

// Resultado consolida os números de uma execução.
type Resultado struct {
	Versao     string
	TotalMs    float64
	EsperaMed  float64 // ms, média de todas as aquisições
	EsperaMax  float64 // ms, maior espera observada
	MaxComendo int32
	Ciclos     [N]int
}

func (m *Metricas) Consolida(versao string) Resultado {
	r := Resultado{Versao: versao, TotalMs: ms(m.total), MaxComendo: m.maxComendo, Ciclos: m.ciclos}
	var soma time.Duration
	var n int
	var max time.Duration
	for i := 0; i < N; i++ {
		soma += m.somaEspera[i]
		n += m.ciclos[i]
		if m.maxEspera[i] > max {
			max = m.maxEspera[i]
		}
	}
	if n > 0 {
		r.EsperaMed = ms(soma) / float64(n)
	}
	r.EsperaMax = ms(max)
	return r
}

func ms(d time.Duration) float64 { return math.Round(float64(d)/1e4) / 100 }

// Imprime o resultado legível e uma linha CSV (usada pelo script de experimentos).
func (r Resultado) Imprime() {
	fmt.Println(strings.Repeat("-", 52))
	fmt.Printf("Versão %s | %d filósofos x %d ciclos | pensa %v, come %v\n", r.Versao, N, CICLOS, TempoPensa, TempoCome)
	fmt.Printf("Tempo total de execução ........ %9.2f ms\n", r.TotalMs)
	fmt.Printf("Tempo médio de espera (garfos) . %9.2f ms\n", r.EsperaMed)
	fmt.Printf("Maior tempo de espera .......... %9.2f ms\n", r.EsperaMax)
	fmt.Printf("Máx. comendo simultaneamente ... %9d\n", r.MaxComendo)
	fmt.Printf("Ciclos por filósofo ............ %v\n", r.Ciclos)
	fmt.Println(strings.Repeat("-", 52))
	c := r.Ciclos
	fmt.Printf("CSV,%s,%.2f,%.2f,%.2f,%d,%d,%d,%d,%d,%d\n",
		r.Versao, r.TotalMs, r.EsperaMed, r.EsperaMax, r.MaxComendo, c[0], c[1], c[2], c[3], c[4])
}

// log opcional de eventos (flag -log), útil para as capturas de tela.
var verbose bool
