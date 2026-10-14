// main.go — escolhe a versão pela linha de comando.
//
//	go run . -v 0            V0 sequencial
//	go run . -v 1            V1 (vigia imprime a tabela do deadlock se travar)
//	go run . -v 1 -pausa 5ms V1 com janela entre os garfos (deadlock garantido)
//	go run . -v 1 -semvigia  V1 deixando o runtime do Go acusar o deadlock
//	go run . -v 2            V2 sincronização excessiva
//	go run . -v 3            V3 ordenação hierárquica de recursos
//	go run . -v 4            V3b variante par/ímpar
//	acrescente -log para ver o rastro em colunas (T = pensando, E = comendo)
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

func main() {
	v := flag.Int("v", 3, "versão: 0, 1, 2, 3 ou 4 (=V3b)")
	pausa := flag.Duration("pausa", 0, "V1: pausa entre pegar o garfo esquerdo e o direito")
	semVigia := flag.Bool("semvigia", false, "V1: desliga o vigia (o runtime do Go detecta o deadlock)")
	flag.BoolVar(&verbose, "log", false, "imprime o rastro de eventos em colunas")
	flag.Parse()

	fmt.Printf("Go %s | %s/%s | GOMAXPROCS=%d | NumCPU=%d\n",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), runtime.NumCPU())
	if verbose {
		fmt.Println("\n[F0]  [F1]  [F2]  [F3]  [F4]")
	}

	var r Resultado
	switch *v {
	case 0:
		r = V0()
	case 1:
		if *pausa > 0 {
			fmt.Printf("V1 com pausa de %v entre os garfos\n", *pausa)
		}
		r = V1(*pausa, !*semVigia)
	case 2:
		r = V2()
	case 3:
		r = V3(false)
	case 4:
		r = V3(true) // V3b: variante par/ímpar
	default:
		fmt.Println("versão inválida")
		os.Exit(1)
	}
	r.Imprime()
}
