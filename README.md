# Major Pick'ems Simulator

This is a basic Python script to simulate tournament stage outcomes for Counter-Strike major tournaments, used to assist decision making for pick'ems. The swiss system follows the seeding rules and format [documented by Valve](https://github.com/ValveSoftware/counter-strike/blob/main/major-supplemental-rulebook.md#seeding), and the tournament rounds are progressed with randomised match outcomes. Each team's ranking from various sources is aggregated to approximate a win probability for each head to head match up. This is by no means an exhaustive or accurate analysis but may give insight to some teams which have higher probability of facing weaker teams to get their 3 wins, or vice versa.

### Command line interface

```
usage: python simulate.py [-h] -f F [-n N] [-k K]

options:
  -h, --help  show this help message and exit
  -f F        path to input data (.json)
  -n N        number of iterations to run
  -k K        number of cores to use
```

### JSON data format

```
{
    "systems": {
        <system name>: <transfer function>
    },
    "sigma": {
        <system name>: <standard deviation for rating>
    },
    "teams": {
        <team name>: {
            "seed": <initial seeding>,
            <system name>: <system rating>
        }
    }
}
```

### Sample output

```text
RESULTS FROM 1,000,000 TOURNAMENT SIMULATIONS

Most likely to 3-0:
1.  Cloud9           27.0%
2.  ENCE             25.4%
3.  HEROIC           22.8%
4.  Apeks            21.1%
5.  Eternal Fire     20.7%
6.  FURIA            20.0%
7.  SAW              12.3%
8.  9Pandas           9.7%
9.  The MongolZ       7.3%
10. AMKAL             5.6%
11. Imperial          5.4%
12. Lynn Vision       5.2%
13. KOI               5.0%
14. ECSTATIC          4.8%
15. paiN              4.4%
16. Legacy            3.3%

Most likely to 3-1 or 3-2:
1.  Cloud9           52.4%
2.  Eternal Fire     52.1%
3.  ENCE             51.9%
4.  HEROIC           51.0%
5.  Apeks            50.5%
6.  FURIA            48.8%
7.  SAW              40.5%
8.  9Pandas          36.7%
9.  The MongolZ      35.2%
10. Imperial         29.1%
11. Lynn Vision      28.3%
12. ECSTATIC         27.9%
13. AMKAL            26.6%
14. paiN             25.0%
15. KOI              24.4%
16. Legacy           19.4%

Most likely to 0-3:
1.  Legacy           25.6%
2.  paiN             21.0%
3.  ECSTATIC         20.0%
4.  KOI              19.2%
5.  Lynn Vision      18.8%
6.  Imperial         18.6%
7.  AMKAL            17.5%
8.  The MongolZ      14.2%
9.  9Pandas          11.1%
10. SAW               9.4%
11. FURIA             4.7%
12. Eternal Fire      4.5%
13. Apeks             4.4%
14. HEROIC            4.1%
15. ENCE              3.7%
16. Cloud9            3.2%

Run time: 17.15 seconds
```
