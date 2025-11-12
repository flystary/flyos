package routing

// DSL
/*
route add static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
route set static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
route delete static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
route add bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100 ] }
route add ospf { prefix 192.168.10.0/24; area 0.0.0.0; type external }
route set pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }
route delete pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }
routes sync {
  static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
  bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100 ] }
  ospf { prefix 192.168.10.0/24; area 0.0.0.0; type external }
  pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }
  static { prefix 20.0.0.0/24; via 192.168.2.1; dev eth1; track yes }
}
*/

/*
add_static_route { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
set_static_route { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
set_static_route { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
*/
