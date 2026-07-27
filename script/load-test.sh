#!/bin/bash

URL="http://localhost:8080/result?roll=100001&reg=202310001&exam_year=2025"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_FILE="$DIR/load_test_results.json"

# Test configurations: "concurrency:requests"
TESTS=(
  "50:10000"
  "100:20000"
  "200:50000"
  "300:75000"
  "500:100000"
  "750:150000"
  "1000:200000"
)

echo ""
echo "=================================================="
echo "  SSC Result Lookup — Load Test"
echo "  URL: $URL"
echo "=================================================="
echo ""

# Check hey is installed
if ! command -v hey &> /dev/null; then
  echo "Error: 'hey' is not installed."
  echo "Install it with: go install github.com/rakyll/hey@latest"
  exit 1
fi

results="["
first=true

for test in "${TESTS[@]}"; do
  concurrency="${test%%:*}"
  requests="${test##*:}"

  echo "--------------------------------------------------"
  echo "  Concurrency: $concurrency | Requests: $requests"
  echo "--------------------------------------------------"

  # Run hey and capture output
  raw=$(hey -n "$requests" -c "$concurrency" "$URL" 2>&1)

  # Parse values from hey output
  rps=$(echo "$raw" | grep "Requests/sec:" | awk '{print $2}')
  avg=$(echo "$raw" | grep "Average:" | awk '{print $2}')
  fastest=$(echo "$raw" | grep "Fastest:" | awk '{print $2}')
  slowest=$(echo "$raw" | grep "Slowest:" | awk '{print $2}')
  p50=$(echo "$raw" | grep "50% in" | awk '{print $3}')
  p75=$(echo "$raw" | grep "75% in" | awk '{print $3}')
  p90=$(echo "$raw" | grep "90% in" | awk '{print $3}')
  p95=$(echo "$raw" | grep "95% in" | awk '{print $3}')
  p99=$(echo "$raw" | grep "99% in" | awk '{print $3}')
  status_200=$(echo "$raw" | grep "\[200\]" | awk '{print $2}')
  total_errors=$((requests - ${status_200:-0}))

  # Print structured summary to terminal
  echo "  RPS:       $rps"
  echo "  Average:   ${avg}s"
  echo "  Fastest:   ${fastest}s"
  echo "  Slowest:   ${slowest}s"
  echo "  p50:       ${p50}s"
  echo "  p75:       ${p75}s"
  echo "  p90:       ${p90}s"
  echo "  p95:       ${p95}s"
  echo "  p99:       ${p99}s"
  echo "  200s:      ${status_200:-0} / $requests"
  echo "  Errors:    $total_errors"
  echo ""

  # Build JSON entry
  if [ "$first" = true ]; then
    first=false
  else
    results="$results,"
  fi

  results="$results
  {
    \"concurrency\": $concurrency,
    \"requests\": $requests,
    \"rps\": $rps,
    \"latency\": {
      \"average_s\": $avg,
      \"fastest_s\": $fastest,
      \"slowest_s\": $slowest,
      \"p50_s\": $p50,
      \"p75_s\": $p75,
      \"p90_s\": $p90,
      \"p95_s\": $p95,
      \"p99_s\": $p99
    },
    \"responses\": {
      \"total\": $requests,
      \"success_200\": ${status_200:-0},
      \"errors\": $total_errors
    }
  }"
done

results="$results
]"

echo "=================================================="
echo "  Writing results to $OUTPUT_FILE"
echo "=================================================="

echo "$results" > "$OUTPUT_FILE"
echo ""
echo "Done. Results saved to $OUTPUT_FILE"
echo ""