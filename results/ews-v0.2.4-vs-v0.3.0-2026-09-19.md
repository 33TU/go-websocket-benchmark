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

## Balanced power profile

The same pairs with the laptop switched from the performance profile to
balanced, 30 seconds of rest before each pair, second pair reversed.

| pair | order | version | echo TPS | echo TP99 | echo CPU | echo memory | rate EER | rate CPU | rate memory |
|---|---|---|---|---|---|---|---|---|---|
| 1 | first | v0.2.4 | 564,539 | 23.97ms | 429% | 135 MB | 4,963 | 401% | 199 MB |
| 1 | second | v0.3.0 | 563,545 | 24.98ms | 427% | 137 MB | 4,876 | 408% | 213 MB |
| 2 | first | v0.3.0 | 563,215 | 25.02ms | 426% | 135 MB | 4,879 | 408% | 201 MB |
| 2 | second | v0.2.4 | 562,880 | 24.75ms | 427% | 132 MB | 4,833 | 411% | 198 MB |

Echo TPS spreads 0.3 percent across all four runs and the rate EER 2.6
percent, against 8 and 11 percent in the performance profile, and the
first-runner advantage is gone. The absolute figures are lower, echo by 5
to 10 percent and the rate EER by about 30 percent, because the sustained
clock is lower and the same messages cost more CPU time (400 percent
against 300). For comparing two builds on this laptop, balanced is the
profile to use; for absolute numbers, the pinned desktop run.

## v0.4.0, same method

v0.4.0 adds events.Dial, Config.UserData and the doc pass on top of v0.3.0;
nothing on the message path changed. Balanced profile, two pairs, the
second reversed, 30 seconds of rest before each pair.

| pair | order | version | echo TPS | echo TP99 | echo CPU | echo memory | rate EER | rate CPU | rate memory |
|---|---|---|---|---|---|---|---|---|---|
| 1 | first | v0.2.4 | 553,019 | 25.20ms | 427% | 141 MB | 4,926 | 404% | 339 MB |
| 1 | second | v0.4.0 | 558,287 | 24.97ms | 428% | 139 MB | 4,909 | 405% | 213 MB |
| 2 | first | v0.4.0 | 556,973 | 25.56ms | 430% | 137 MB | 4,875 | 408% | 203 MB |
| 2 | second | v0.2.4 | 554,925 | 25.04ms | 427% | 141 MB | 4,826 | 412% | 235 MB |

Echo TPS within 1 percent across all four runs and rate EER within 2,
with no order effect: the same result as v0.3.0. The rate-test memory
average is the sampled column that swings with GC timing, as the 339 MB
outlier for the same v0.2.4 binary that measured 199 MB earlier shows.

## gws v1.10.2 against ews v0.4.0, same method

Two pairs, balanced profile, the second reversed.

| pair | order | server | echo TPS | echo TP99 | echo CPU | echo memory | rate EER | rate CPU | rate memory | dropped |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | first | gws | 547,334 | 26.25ms | 395% | 169 MB | 3,225 | 379% | 234 MB | 0 |
| 1 | second | ews v0.4.0 | 548,675 | 24.37ms | 393% | 143 MB | 4,900 | 406% | 201 MB | 0 |
| 2 | first | ews v0.4.0 | 556,050 | 24.82ms | 428% | 133 MB | 4,875 | 408% | 259 MB | 0 |
| 2 | second | gws | 544,531 | 26.42ms | 393% | 173 MB | 3,299 | 394% | 235 MB | 223,358 |

Echo ties, as always. ews holds 20 percent less memory during the echo
test. In the rate test ews answers 1.5 times as many echoes per CPU point
and took every offered message in both runs; gws dropped 223,358 of 19.9M
in its second run. The ratio is 1.5 here against 1.8 on the desktop
because the balanced profile's lower clock leaves less headroom for the
coalescing to convert into throughput.
