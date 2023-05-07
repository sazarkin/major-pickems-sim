# major-pickems-sim

This is a basic Python script to simulate tournament stage outcomes for CS:GO major tournaments, used to assist decision making for tournament pickems. The swiss system is initialised given each team's actual seed, and the tournament rounds are progressed with randomised match outcomes. Each team's ranking on HLTV, ESL, and GosuGamers is aggregated to approximate a win probability for each head to head match up. This is by no means an exhaustive or accurate analysis but may give insight to some teams which have higher probability of facing weaker teams to get their 3 wins, or vice versa.

Sample output:

```text
RESULTS FROM 1,000,000 TOURNAMENT SIMULATIONS

Most likely to advance:
1.  FaZe         88.32%
2.  G2           87.75%
3.  Liquid       71.33%
4.  ENCE         60.63%
5.  FORZE        54.75%
6.  MOUZ         52.36%
7.  OG           52.08%
8.  NiP          51.87%
9.  paiN         50.35%
10. Complexity   46.64%
11. Monte        32.76%
12. GamerLegion  31.37%
13. TheMongolz   31.14%
14. Fluxo        31.03%
15. Apeks        28.81%
16. Grayhound    28.81%

Most likely to 3-0:
1.  G2           33.49%
2.  FaZe         32.02%
3.  Liquid       22.33%
4.  ENCE         15.05%
5.  MOUZ         12.12%
6.  FORZE        12.06%
7.  OG           11.41%
8.  NiP          11.15%
9.  paiN         10.25%
10. Complexity   9.45%
11. Monte        5.39%
12. GamerLegion  5.37%
13. Fluxo        5.27%
14. Grayhound    5.17%
15. TheMongolz   4.82%
16. Apeks        4.65%

Most likely to 0-3:
1.  Apeks        21.95%
2.  Grayhound    21.52%
3.  TheMongolz   20.87%
4.  Fluxo        19.5%
5.  GamerLegion  18.89%
6.  Monte        18.8%
7.  Complexity   11.87%
8.  NiP          10.94%
9.  OG           10.84%
10. MOUZ         10.24%
11. paiN         9.47%
12. FORZE        9.0%
13. ENCE         7.86%
14. Liquid       5.01%
15. FaZe         1.68%
16. G2           1.57%

Run time: 6.853 seconds
```
