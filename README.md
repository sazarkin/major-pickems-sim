This is a basic Python script to simulate tournament stage outcomes for CS:GO major tournaments, used to assist decision making for tournament pickems. The swiss system is initialised given each team's actual seed, and the tournament rounds are progressed with randomised match outcomes. Each team's ranking on HLTV, ESL, and GosuGamers is aggregated to approximate a win probability for each head to head match up. This is by no means an exhaustive or accurate analysis but may give insight to some teams which have higher probability of facing weaker teams to get their 3 wins, or vice versa.

Sample output:
```
RESULTS FROM 1,000,000 TOURNAMENT SIMULATIONS

Most likely to advance:
1.  FaZe         76.97%
2.  Liquid       72.54%
3.  NaVi         71.83%
4.  Vitality     66.65%
5.  Heroic       57.57%
6.  NiP          53.81%
7.  Outsiders    51.25%
8.  Spirit       49.6%
9.  MOUZ         49.47%
10. ENCE         46.49%
11. FURIA        44.88%
12. Cloud9       44.15%
13. BIG          39.46%
14. fnatic       27.72%
15. Sprout       24.2%
16. BNE          23.4%

Most likely to 3-0:
1.  FaZe         26.4%
2.  NaVi         22.18%
3.  Liquid       21.72%
4.  Vitality     18.81%
5.  Heroic       14.06%
6.  NiP          13.16%
7.  Spirit       11.91%
8.  Outsiders    11.01%
9.  MOUZ         10.39%
10. Cloud9       10.34%
11. ENCE         9.82%
12. FURIA        9.06%
13. BIG          8.06%
14. fnatic       4.81%
15. Sprout       4.28%
16. BNE          4.0%

Most likely to 0-3:
1.  BNE          25.0%
2.  Sprout       24.77%
3.  fnatic       22.02%
4.  BIG          14.35%
5.  FURIA        13.69%
6.  ENCE         13.47%
7.  Cloud9       12.52%
8.  MOUZ         12.38%
9.  Outsiders    11.28%
10. Spirit       10.86%
11. NiP          10.31%
12. Heroic       9.64%
13. Vitality     5.76%
14. Liquid       5.21%
15. NaVi         4.92%
16. FaZe         3.81%

Run time: 9.331 seconds
```
