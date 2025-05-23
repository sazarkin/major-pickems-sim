# Major Pick'ems Simulator

This is a basic Python script to simulate tournament stage outcomes for Counter-Strike major tournaments, used to assist decision making for pick'ems. The swiss system follows the seeding rules and format [documented by Valve](https://github.com/ValveSoftware/counter-strike/blob/main/major-supplemental-rulebook.md#seeding), and the tournament rounds are progressed with randomised match outcomes. Each team's ranking from various sources is aggregated to approximate a win probability for each head to head match up. This is by no means an exhaustive or accurate analysis but may give insight to some teams which have higher probability of facing weaker teams to get their 3 wins, or vice versa.

### Command line interface

```
usage: python simulate.py [-h] -f F [-n N] [-k K] [-p P] [-s S]

options:
  -h, --help  show this help message and exit
  -f F        path to input data (.json)
  -n N        number of iterations to run
  -k K        number of cores to use
  -p P        number of predictions to run
  -s S        random seed
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
1.  FaZe             38.4%
2.  Spirit           31.4%
3.  Vitality         31.4%
4.  MOUZ             25.6%
5.  Virtus.pro       18.4%
6.  Natus Vincere    15.7%
7.  G2               14.2%
8.  Complexity        6.5%
9.  Cloud9            4.6%
10. HEROIC            3.7%
11. Eternal Fire      3.5%
12. FURIA             2.5%
13. The MongolZ       1.3%
14. Imperial          1.0%
15. ECSTATIC          0.9%
16. paiN              0.8%

Most likely to 3-1 or 3-2:
1.  G2               58.7%
2.  MOUZ             58.1%
3.  Natus Vincere    58.0%
4.  Vitality         57.5%
5.  Virtus.pro       57.0%
6.  Spirit           56.4%
7.  FaZe             53.1%
8.  Complexity       38.9%
9.  Cloud9           36.4%
10. HEROIC           31.6%
11. Eternal Fire     30.8%
12. FURIA            21.4%
13. The MongolZ      12.9%
14. Imperial         10.1%
15. ECSTATIC          9.9%
16. paiN              9.1%

Most likely to 0-3:
1.  Imperial         31.8%
2.  ECSTATIC         31.5%
3.  paiN             31.0%
4.  The MongolZ      29.1%
5.  FURIA            21.3%
6.  Eternal Fire     13.2%
7.  HEROIC           12.4%
8.  Cloud9           10.2%
9.  Complexity        7.2%
10. G2                2.7%
11. Virtus.pro        2.6%
12. Natus Vincere     2.5%
13. MOUZ              1.4%
14. Spirit            1.2%
15. Vitality          1.0%
16. FaZe              0.8%

Run time: 17.70 seconds
```
