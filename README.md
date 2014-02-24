go_perf
=======

Simple linux performance monitor.

It collects cpu info, memory info and top processes list.
Each metric is being collected in its own routine with its own period:
- cpu usage - every 1 second 
- memory usage - every 1 second
- top processes - every 5 seconds
Collected data sends in udp packet to a server, that collects it and stores database
