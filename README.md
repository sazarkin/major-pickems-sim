# Major Pick'ems Simulator

This is a Go program to simulate tournament stage outcomes for Counter-Strike major tournaments, used to assist decision making for pick'ems. The swiss system follows the seeding rules and format [documented by Valve](https://github.com/ValveSoftware/counter-strike/blob/main/major-supplemental-rulebook.md#seeding), and the tournament rounds are progressed with randomised match outcomes. Each team's ranking from various sources is aggregated to approximate a win probability for each head to head match up. This is by no means an exhaustive or accurate analysis but may give insight to some teams which have higher probability of facing weaker teams to get their 3 wins, or vice versa.

### Building the executable

First, compile the program to create an executable binary:

```bash
go build -o simulate simulate.go swisssystem.go
```

This will generate a `simulate` binary (or `simulate.exe` on Windows) that you can run directly.

### Running the simulation

Once built, use the binary with the desired options:

```bash
./simulate -f <data.json> [options]
```

#### Command line options

```
  -f string
        path to input data (.json)
  -k int
        number of cores to use (default 24)
  -n int
        number of iterations to run (default 1000000)
  -p int
        number of predictions to run (default 1000)
  -profile string
        write cpu profile to file
  -s int
        random seed
```

#### Example

```bash
./simulate -f data/2025_budapest_stage3.json
```

### JSON data format

```json
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

After running the simulation, you'll see the top 5 predictions with their success rates:

```bash
➜  ./simulate -f data/2025_budapest_stage3.json
Percent of success: 27.98%
69d4f
'3-0': MOUZ, Team Falcons
'3-1 or 3-2': Team Vitality, The MongolZ, Team Spirit, Team Liquid, FURIA, G2 Esports
'0-3': Passion UA, Imperial Esports

Percent of success: 27.91%
1d83e
'3-0': FURIA, Team Vitality
'3-1 or 3-2': Team Liquid, Team Spirit, Natus Vincere, The MongolZ, Team Falcons, MOUZ
'0-3': 3DMAX, Imperial Esports

Percent of success: 27.80%
7ae19
'3-0': FURIA, Natus Vincere
'3-1 or 3-2': Team Vitality, Team Liquid, Team Falcons, Team Spirit, MOUZ, G2 Esports
'0-3': Imperial Esports, Passion UA

Percent of success: 27.55%
7f9e6
'3-0': FURIA, Team Vitality
'3-1 or 3-2': The MongolZ, G2 Esports, MOUZ, Team Falcons, Team Liquid, Team Spirit
'0-3': 3DMAX, Imperial Esports

Percent of success: 27.47%
5f85f
'3-0': FURIA, Team Falcons
'3-1 or 3-2': Team Spirit, G2 Esports, The MongolZ, Team Vitality, Natus Vincere, paiN Gaming
'0-3': Imperial Esports, Passion UA
```

The output shows:
1. The success percentage for each prediction
2. A 5-character hash identifying the prediction
3. Teams predicted to go 3-0, 3-1/3-2, and 0-3
