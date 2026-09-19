# ews v0.2.4 against v0.3.0, 10k connections

Run on 2026-09-19 on the laptop (13th Gen Intel Core i9-13900H, Go 1.27.0,
Linux 6.12), servers pinned to cores 0 to 5 and the client to 6 to 19,
after several hours of other benchmarking on the same machine. Three
rounds, alternating; round 3 with the order reversed.

```sh
taskset -c 0-5 ./output/bin/ews-$v.server &
taskset -c 6-19 ./output/bin/bench.client -f=ews -c=10000 -dc=2000 -en=2000000 -rate=true -rd=10
```

v0.3.0 changes the transport to a net.Conn, decodes frame headers in place
(14 to 27 percent faster small reads in ews's own benchmarks) and adds the
events package; the server here is the same Queue-based echo as before.

| round | order | version | echo TPS | echo TP99 | echo CPU | echo memory | rate EER | rate CPU | rate memory |
|---|---|---|---|---|---|---|---|---|---|
| 1 | first | v0.2.4 | 634,445 | 20.96ms | 433% | 145 MB | 7,017 | 284% | 172 MB |
| 1 | second | v0.3.0 | 587,662 | 27.72ms | 432% | 143 MB | 6,698 | 297% | 182 MB |
| 2 | first | v0.2.4 | 589,159 | 23.26ms | 433% | 137 MB | 6,479 | 307% | 243 MB |
| 2 | second | v0.3.0 | 585,154 | 23.14ms | 425% | 135 MB | 6,283 | 317% | 178 MB |
| 3 | first | v0.3.0 | 624,230 | 21.93ms | 429% | 139 MB | 6,812 | 292% | 171 MB |
| 3 | second | v0.2.4 | 604,374 | 22.32ms | 430% | 139 MB | 6,581 | 302% | 202 MB |

Whichever version runs first in a round wins by 3 to 4 percent on both
tests, and the rate EER declines monotonically with wall-clock time
(7,017, 6,698, 6,479, 6,283) until the cool-down before round 3. That is
the machine warming, not the code: the two versions are equal within
noise on throughput, CPU and memory, and every run received all 19.9M
rate-test messages with no drops. The per-frame gains v0.3.0 shows in
the Go benchmarks are nanoseconds against a microsecond syscall, so this
is the expected result; the desktop run with the full field is the one
to quote.
