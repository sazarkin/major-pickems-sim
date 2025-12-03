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

### Alternative: running without building

If you prefer not to build a binary first, you can still use `go run` with both source files:

```bash
go run simulate.go swisssystem.go -f data/2025_budapest_stage3.json
```

However, building the binary first is more efficient for multiple runs.

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

After running the simulation, you'll see the top 5 predictions with their success rates:

```text
Percent of success: 26.00%
dea74
'0-3': Passion UA, Imperial Esports
'3-0': Team Vitality, Team Spirit
'3-1 or 3-2': G2 Esports, MOUZ, B8, FURIA, Natus Vincere, The MongolZ

Percent of success: 22.36%
f2a85
'3-0': G2 Esports, FURIA
'3-1 or 3-2': paiN Gaming, MOUZ, Team Liquid, The MongolZ, Natus Vincere, B8
'0-3': Passion UA, Imperial Esports

Percent of success: 22.25%
020d9
'0-3': Imperial Esports, PARIVISION
'3-0': Natus Vincere, Team Vitality
'3-1 or 3-2': Team Liquid, MOUZ, G2 Esports, Team Falcons, FaZe Clan, FURIA

Percent of success: 21.92%
45817
'3-0': Imperial Esports, Team Falcons
'3-1 or 3-2': MOUZ, G2 Esports, Team Liquid, Team Vitality, Natus Vincere, FURIA
'0-3': Passion UA, FaZe Clan

Percent of success: 21.89%
47f75
'3-0': Team Falcons, Passion UA
'3-1 or 3-2': FURIA, MOUZ, The MongolZ, Team Spirit, Natus Vincere, Team Vitality
'0-3': PARIVISION, FaZe Clan
```

The output shows:
1. The success percentage for each prediction
2. A 5-character hash identifying the prediction
3. Teams predicted to go 3-0, 3-1/3-2, and 0-3
