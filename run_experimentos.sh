#!/usr/bin/env bash
# run_experimentos.sh — roda todo o experimento do T1 e salva em resultados/.
# Uso: bash run_experimentos.sh        (leva ~2 minutos)
set -u
cd "$(dirname "$0")"
EXEC=${EXEC:-5}          # execuções por versão (enunciado: mínimo 5)
OUT=resultados
mkdir -p "$OUT"

echo "== Compilando =="
go build -o filosofos . || exit 1
echo "comando: go build -o filosofos .  &&  ./filosofos -v <versao>" > "$OUT/comandos.txt"

echo "== Ambiente =="
{
  go version
  grep PRETTY_NAME /etc/os-release
  lscpu | grep -E "Model name|^CPU\(s\)"
  uname -m
} | tee "$OUT/ambiente.txt"

echo "== Versões que concluem: $EXEC execuções cada =="
echo "versao,execucao,total_ms,espera_media_ms,espera_max_ms,max_comendo,c0,c1,c2,c3,c4" > "$OUT/resultados.csv"
for v in 0 2 3 4; do
  for i in $(seq "$EXEC"); do
    linha=$(./filosofos -v $v | grep '^CSV,' | cut -d, -f2-)
    nome=${linha%%,*}; resto=${linha#*,}
    echo "$nome,$i,$resto" | tee -a "$OUT/resultados.csv"
  done
done

echo "== V1 sem pausa: quantas de 20 execuções travam? =="
ok=0; dl=0
for i in $(seq 20); do
  timeout 30 ./filosofos -v 1 > "$OUT/v1_sem_pausa_$i.txt" 2>&1
  if [ $? -eq 0 ]; then ok=$((ok+1)); else dl=$((dl+1)); fi
done
echo "V1 sem pausa: 20 execuções, $ok concluíram, $dl travaram em deadlock" | tee "$OUT/v1_frequencia.txt"

echo "== V1 com pausa de 5ms entre os garfos (evidência do deadlock) =="
timeout 30 ./filosofos -v 1 -pausa 5ms -log > "$OUT/v1_deadlock_vigia.txt" 2>&1
tail -20 "$OUT/v1_deadlock_vigia.txt"

echo "== V1 sem vigia: o runtime do Go acusa o deadlock =="
timeout 30 ./filosofos -v 1 -pausa 5ms -semvigia > "$OUT/v1_deadlock_runtime.txt" 2>&1
head -5 "$OUT/v1_deadlock_runtime.txt"

echo "== Rastros em colunas para as capturas de tela =="
for v in 0 2 3; do ./filosofos -v $v -log > "$OUT/rastro_v$v.txt" 2>&1; done

echo
echo "Pronto. Me mande o conteúdo de: $OUT/resultados.csv, $OUT/v1_frequencia.txt,"
echo "$OUT/v1_deadlock_vigia.txt (final) e $OUT/v1_deadlock_runtime.txt (início)."
