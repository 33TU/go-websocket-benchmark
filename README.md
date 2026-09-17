# go-websocket-benchmark

This fork adds [ews](https://github.com/33TU/ews) as a framework, in two
shapes, and tracks current library versions: gws v1.10.2, gorilla v1.5.3,
sonic bumped so hertz builds under Go 1.27.

- `ews`: `transport.Server` without net/http, echo through a `Queue`; the
  counterpart of gws with `WriteAsync`.
- `ews_sync`: net/http, synchronous `Write`; the counterpart of `gws_std`.

## Results

- [ews against gws, 10k connections, laptop](results/ews-vs-gws-2026-09-16.md):
  i9-13900H. Echo ties; in the rate test ews takes the whole offered load
  with no drops at 1.8 times gws's echoes per CPU point, while gws drops.
- [ews against gws, 10k connections, desktop](results/ews-vs-gws-2026-09-17-9950x3d.md):
  Ryzen 9 9950X3D. Both take the whole load; ews does it on 193 percent CPU
  against 347, the same 1.8 ratio.
- [The full suite, 10k and 30k connections, desktop](results/suite-2026-09-17-9950x3d.md):
  every framework in `script/config.sh`, echo and rate, plus an echo cell
  at 256 KiB payloads, with the generated reports in
  [suite-2026-09-17-9950x3d/](results/suite-2026-09-17-9950x3d/). At 256 KiB
  every server is bound by loopback bandwidth and the difference is CPU:
  ews moves the same bytes on 180 percent where gws needs 300.

Two things learned running it. The echo test keeps one request in flight
per connection, so its TPS is connections divided by round trip and every
good library ties on it; the differences are in the rate test, where the
Queue coalesces a backlog into one writev. And the machine can sit in a
slower state for reasons that do not show in load: run the same server
first and last as a drift check before reading a difference as a library
difference, pin the servers and the client to separate physical cores, and
kill the hertz servers with SIGKILL, since they ignore SIGTERM.

- support 1m-connections client

## before running the test
- make sure setting the correct system env, for example:

```sh
sysctl -w net.ipv4.ip_local_port_range="1024 65535"
sysctl -w fs.file-max=2000500
sysctl -w fs.nr_open=2000500
sysctl -w net.nf_conntrack_max=2000500
ulimit -n 2000500
sysctl -w net.ipv4.tcp_mem='131072  262144  524288'
sysctl -w net.ipv4.tcp_rmem='8760  256960  4088000'
sysctl -w net.ipv4.tcp_wmem='8760  256960  4088000'
sysctl -w net.core.rmem_max=16384
sysctl -w net.core.wmem_max=16384
sysctl -w net.core.somaxconn=2048
sysctl -w net.ipv4.tcp_max_syn_backlog=2048
sysctl -w /proc/sys/net/core/netdev_max_backlog=2048
# sysctl -w net.ipv4.tcp_tw_recycle=1 # client nat tcp-handshak problem
sysctl -w net.ipv4.tcp_tw_reuse=1
```

## nbio 1m-connections-benchmark


run:
```sh
git clone https://github.com/33TU/go-websocket-benchmark.git
cd go-websocket-benchmark
./script/1m_conns_benchmark.sh
```

here is the result on my ubuntu vm:
```sh
--------------------------------------------------------------
BenchType  : Connections
Framework  : nbio_nonblocking
Connections: 1000000
Concurrency: 5000
Success    : 1000000
Failed     : 0
Used       : 41.56s
TPS        : 24061
Min        : 20ns
Avg        : 192.57ms
Max        : 41.52s
TP50       : 30ns
TP75       : 30ns
TP90       : 30ns
TP95       : 31ns
TP99       : 31ns
--------------------------------------------------------------
BenchType  : BenchEcho
Framework  : nbio_nonblocking
Conns      : 1000000
Concurrency: 50000
Payload    : 1024
Total      : 5000000
Success    : 5000000
Failed     : 0
Used       : 47.02s
CPU Min    : 0.00%
CPU Avg    : 340.08%
CPU Max    : 386.93%
MEM Min    : 1.76G
MEM Avg    : 1.91G
MEM Max    : 1.94G
TPS        : 106348
Min        : 436.16us
Avg        : 465.78ms
Max        : 2.42s
TP50       : 412.36ms
TP75       : 600.92ms
TP90       : 779.92ms
TP95       : 1.04s
TP99       : 1.35s
--------------------------------------------------------------------------
```

## benchmark for all frameworks
run:
```sh
git clone https://github.com/33TU/go-websocket-benchmark.git
cd go-websocket-benchmark
./script/benchmarkN.sh

# if you want to change the benchmark config, just read the script and edit:
# go-websocket-benchmark/script/config.sh
```

Results from 2026-09-17 on a Ryzen 9 9950X3D, servers on six cores, 10,000
connections, 1024-byte payloads; the full run, with 30k connections too, is
in [results/](results/suite-2026-09-17-9950x3d.md).

[BenchEcho] Report

|    Framework     |   TPS   |   EER   |   Min   |   Avg   |   Max   |  TP50   |  TP75   |  TP90   |  TP95   |  TP99   | Used  |  Total  | Success | Failed | Conns | Concurrency | Payload | CPU Min | CPU Avg | CPU Max | MEM Min | MEM Avg | MEM Max |
|     ---          |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |  ---  |   ---   |   ---   |  ---   |  ---  |     ---     |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |
|   fasthttp       | 1072935 | 2763.18 | 30.32us | 9.28ms  | 21.27ms | 8.88ms  | 9.35ms  | 11.95ms | 12.35ms | 13.70ms | 1.86s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 360.65  | 388.30  | 415.95  | 239.87M | 239.87M | 239.87M |
|    gobwas        | 691324  | 1930.51 | 10.07us | 14.42ms | 75.96ms | 13.63ms | 15.70ms | 17.34ms | 17.91ms | 26.94ms | 2.89s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 329.69  | 358.10  | 373.66  | 262.82M | 262.82M | 262.82M |
|    gorilla       | 1063168 | 2748.46 | 21.91us | 9.37ms  | 26.15ms | 8.87ms  | 10.53ms | 12.00ms | 12.45ms | 14.78ms | 1.88s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 363.74  | 386.82  | 409.90  | 239.50M | 239.50M | 239.50M |
|     gws          | 1097026 | 2950.58 | 16.29us | 9.07ms  | 36.26ms | 8.49ms  | 9.69ms  | 11.67ms | 12.06ms | 14.29ms | 1.82s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 338.67  | 371.80  | 404.93  | 181.77M | 181.77M | 181.77M |
|    gws_std       | 1159106 | 2942.82 | 12.59us | 8.59ms  | 20.49ms | 7.88ms  | 8.72ms  | 11.46ms | 11.72ms | 13.08ms | 1.73s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 393.88  | 393.88  | 393.88  | 159.62M | 159.62M | 159.62M |
|    hertz         | 663479  | 2062.93 | 20.52us | 15.01ms | 54.20ms | 13.46ms | 14.57ms | 24.43ms | 25.93ms | 29.25ms | 3.01s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 319.95  | 321.62  | 322.95  | 518.61M | 535.03M | 551.46M |
|   hertz_std      | 1001382 | 2377.10 | 19.82us | 9.95ms  | 24.95ms | 9.56ms  | 10.33ms | 12.47ms | 13.01ms | 14.89ms | 2.00s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 418.58  | 421.26  | 423.94  | 346.43M | 346.43M | 346.43M |
|  nbio_blocking   | 1083084 | 2874.55 | 12.96us | 9.19ms  | 24.17ms | 8.58ms  | 10.21ms | 11.82ms | 12.40ms | 14.18ms | 1.85s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 344.71  | 376.78  | 408.85  | 183.56M | 183.56M | 183.56M |
|   nbio_mixed     | 1128597 | 2807.81 | 13.21us | 8.83ms  | 24.18ms | 8.18ms  | 8.78ms  | 11.62ms | 12.12ms | 15.03ms | 1.77s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 401.95  | 401.95  | 401.95  | 320.82M | 320.82M | 320.82M |
| nbio_nonblocking | 968986  | 2422.88 | 9.30us  | 10.28ms | 58.04ms | 9.42ms  | 13.06ms | 17.53ms | 20.39ms | 29.54ms | 2.06s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 399.92  | 399.93  | 399.94  | 121.30M | 121.30M | 121.30M |
|   nbio_std       | 1145238 | 2937.10 | 18.18us | 8.70ms  | 29.76ms | 7.84ms  | 9.27ms  | 11.54ms | 11.99ms | 15.19ms | 1.75s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 389.92  | 389.92  | 389.92  | 187.12M | 187.12M | 187.12M |
|    nettyws       | 1098872 | 2986.75 | 9.03us  | 9.07ms  | 23.22ms | 8.49ms  | 9.81ms  | 11.72ms | 12.17ms | 14.30ms | 1.82s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 333.95  | 367.92  | 401.88  | 170.69M | 170.69M | 170.69M |
|    nhooyr        | 730534  | 1543.13 | 19.90us | 13.63ms | 23.80ms | 13.56ms | 13.92ms | 14.75ms | 15.77ms | 17.85ms | 2.74s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 472.95  | 473.41  | 473.88  | 353.34M | 353.34M | 353.34M |
|    quickws       | 1184546 | 3159.89 | 14.67us | 8.41ms  | 25.92ms | 7.55ms  | 8.54ms  | 11.28ms | 11.56ms | 13.41ms | 1.69s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 374.87  | 374.87  | 374.87  | 132.74M | 132.74M | 132.74M |
|     ews          | 1130573 | 2841.58 | 37.73us | 8.81ms  | 20.06ms | 8.12ms  | 9.03ms  | 11.66ms | 11.97ms | 13.40ms | 1.77s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 397.87  | 397.87  | 397.87  | 145.31M | 145.31M | 145.31M |
|   ews_sync       | 1174460 | 3019.60 | 13.13us | 8.47ms  | 19.67ms | 7.66ms  | 8.48ms  | 11.41ms | 11.72ms | 13.38ms | 1.70s | 2000000 | 2000000 |   0    | 10000 |    10000    |  1024   | 388.95  | 388.95  | 388.95  | 151.27M | 151.27M | 151.27M |

[BenchRate] Report

|    Framework     | Duration | EchoEER  | Packet Sent | Bytes Sent | Packet Recv | Bytes Recv | Conns | SendRate | Payload | CPU Min | CPU Avg | CPU Max | MEM Min | MEM Avg | MEM Max |
|     ---          |   ---    |   ---    |     ---     |    ---     |     ---     |    ---     |  ---  |   ---    |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |   ---   |
|   fasthttp       |  10.00s  | 5772.92  |  19900960   |   18.98G   |  19900960   |   18.98G   | 10000 |   200    |  1024   | 342.96  | 344.73  | 347.92  | 270.62M | 323.83M | 353.01M |
|    gobwas        |  10.00s  | 3002.10  |  11867440   |   11.32G   |  11618691   |   11.08G   | 10000 |   200    |  1024   | 381.71  | 387.02  | 392.95  | 312.68M | 327.30M | 345.15M |
|    gorilla       |  10.00s  | 5755.19  |  19900000   |   18.98G   |  19900000   |   18.98G   | 10000 |   200    |  1024   | 340.72  | 345.78  | 349.94  | 270.32M | 319.19M | 342.40M |
|     gws          |  10.00s  | 5687.02  |  19900000   |   18.98G   |  19900000   |   18.98G   | 10000 |   200    |  1024   | 348.47  | 349.92  | 353.95  | 217.74M | 219.18M | 221.93M |
|    gws_std       |  10.00s  | 5809.76  |  19900000   |   18.98G   |  19900000   |   18.98G   | 10000 |   200    |  1024   | 338.89  | 342.53  | 343.96  | 177.36M | 184.81M | 198.16M |
|    hertz         |  10.00s  | 4329.50  |  16477640   |   15.71G   |  16130392   |   15.38G   | 10000 |   200    |  1024   | 365.04  | 372.57  | 376.95  | 670.29M | 759.70M | 797.17M |
|   hertz_std      |  10.00s  | 5423.65  |  19900010   |   18.98G   |  19900010   |   18.98G   | 10000 |   200    |  1024   | 364.97  | 366.91  | 368.68  | 377.18M | 437.42M | 482.09M |
|  nbio_blocking   |  10.00s  | 5669.24  |  19904200   |   18.98G   |  19904200   |   18.98G   | 10000 |   200    |  1024   | 344.76  | 351.09  | 355.93  | 212.23M | 212.51M | 212.58M |
|   nbio_mixed     |  10.00s  | 5753.07  |  19900340   |   18.98G   |  19900340   |   18.98G   | 10000 |   200    |  1024   | 340.62  | 345.91  | 349.93  | 431.76M | 477.41M | 526.35M |
| nbio_nonblocking |  10.00s  | 4853.78  |  18294650   |   17.45G   |  18158863   |   17.32G   | 10000 |   200    |  1024   | 368.95  | 374.12  | 380.93  | 591.62M | 690.07M | 722.20M |
|   nbio_std       |  10.00s  | 5851.92  |  19996920   |   19.07G   |  19962766   |   19.04G   | 10000 |   200    |  1024   | 338.94  | 341.13  | 344.72  | 200.54M | 202.51M | 204.15M |
|    nettyws       |  10.00s  | 5522.83  |  19900010   |   18.98G   |  19900010   |   18.98G   | 10000 |   200    |  1024   | 357.92  | 360.32  | 361.99  | 212.84M | 224.96M | 236.06M |
|    nhooyr        |  10.00s  | 2921.12  |  12725450   |   12.14G   |  12489582   |   11.91G   | 10000 |   200    |  1024   | 420.96  | 427.56  | 431.94  | 384.12M | 456.43M | 497.32M |
|    quickws       |  10.00s  | 5897.67  |  19900000   |   18.98G   |  19900000   |   18.98G   | 10000 |   200    |  1024   | 330.75  | 337.42  | 340.94  | 141.45M | 146.59M | 147.95M |
|     ews          |  10.00s  | 10222.95 |  19900000   |   18.98G   |  19900000   |   18.98G   | 10000 |   200    |  1024   | 191.78  | 194.66  | 196.78  | 216.53M | 216.53M | 216.53M |
|   ews_sync       |  10.00s  | 6079.41  |  19900000   |   18.98G   |  19900000   |   18.98G   | 10000 |   200    |  1024   | 320.92  | 327.33  | 329.88  | 151.27M | 151.27M | 151.27M |
