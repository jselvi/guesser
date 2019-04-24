read OUT

echo "beef1234cafe
faa123" | grep "$OUT" 2>&1 >/dev/null
echo $?
