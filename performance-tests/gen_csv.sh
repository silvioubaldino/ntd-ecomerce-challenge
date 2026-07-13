#!/usr/bin/env bash
# usage: performance-tests/gen_csv.sh 40000 > performance-tests/test-3-csv-import-batching/import_big.csv
n=${1:-40000}
echo "name,sku,description,category,price,stock,weight_kg"
for i in $(seq 1 "$n"); do
  echo "Product $i,BULK-$i,desc $i,tools,$(( (i%490)+10 )).50,$(( i%100 )),1.250"
done
