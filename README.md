This is a basic Python script to simulate tournament stage outcomes for CS:GO major tournaments, used to assist decision making for tournament pickems. The swiss system is initialised given each team's actual seed, and the tournament rounds are progressed with randomised match outcomes. Each team's ranking on HLTV, ESL, and GosuGamers is aggregated to approximate a win probability for each head to head match up. This is by no means an exhaustive or accurate analysis but may give insight to some teams which have higher probability of facing weaker teams to get their 3 wins, or vice versa.

Sample output:
```
RESULTS FROM 100,000,000 TOURNAMENT SIMULATIONS

Most likely to advanced:
1.  FaZe         77.02%
2.  Liquid       72.55%
3.  NaVi         71.85%
4.  Vitality     66.65%
5.  Heroic       57.59%
6.  NiP          53.85%
7.  Outsiders    51.3%
8.  Spirit       49.61%
9.  MOUZ         49.45%
10. ENCE         46.55%
11. FURIA        44.86%
12. Cloud9       44.06%
13. BIG          39.42%
14. fnatic       27.7%
15. Sprout       24.19%
16. BNE          23.35%

Most likely to 3-0:
1.  FaZe         26.39%
2.  NaVi         22.16%
3.  Liquid       21.76%
4.  Vitality     18.81%
5.  Heroic       14.05%
6.  NiP          13.14%
7.  Spirit       11.9%
8.  Outsiders    11.01%
9.  MOUZ         10.4%
10. Cloud9       10.34%
11. ENCE         9.87%
12. FURIA        9.07%
13. BIG          8.05%
14. fnatic       4.8%
15. Sprout       4.24%
16. BNE          4.03%

Most likely to 0-3:
1.  BNE          24.98%
2.  Sprout       24.76%
3.  fnatic       22.01%
4.  BIG          14.34%
5.  FURIA        13.74%
6.  ENCE         13.49%
7.  Cloud9       12.57%
8.  MOUZ         12.37%
9.  Outsiders    11.27%
10. Spirit       10.91%
11. NiP          10.34%
12. Heroic       9.6%
13. Vitality     5.75%
14. Liquid       5.21%
15. NaVi         4.9%
16. FaZe         3.76%

Run time: 1659.583 seconds
```
