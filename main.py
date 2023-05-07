import random
import multiprocessing
import time
from functools import cache
from pprint import pprint


stat_keys = ["advance", "3-0", "0-3"]
rating_systems = ["hltv", "esl", "gosu"]

team_data = {
    "Monte":        {"seed": 1,     "hltv": 113,   "esl": 182,     "gosu": 1218},
    "paiN":         {"seed": 2,     "hltv": 178,   "esl": 442,     "gosu": 1232},
    "G2":           {"seed": 3,     "hltv": 697,   "esl": 1322,    "gosu": 1553},
    "GamerLegion":  {"seed": 4,     "hltv": 78,    "esl": 107,     "gosu": 1184},
    "FORZE":        {"seed": 5,     "hltv": 195,   "esl": 419,     "gosu": 1240},
    "Apeks":        {"seed": 6,     "hltv": 75,    "esl": 80,      "gosu": 1185},
    "NiP":          {"seed": 7,     "hltv": 216,   "esl": 350,     "gosu": 1262},
    "OG":           {"seed": 8,     "hltv": 239,   "esl": 292,     "gosu": 1293},
    "ENCE":         {"seed": 9,     "hltv": 290,   "esl": 559,     "gosu": 1313},
    "MOUZ":         {"seed": 10,    "hltv": 239,   "esl": 409,     "gosu": 1256},
    "Liquid":       {"seed": 11,    "hltv": 418,   "esl": 634,     "gosu": 1358},
    "Grayhound":    {"seed": 12,    "hltv": 101,   "esl": 95,      "gosu": 1066},
    "Complexity":   {"seed": 13,    "hltv": 161,   "esl": 301,     "gosu": 1158},
    "TheMongolz":   {"seed": 14,    "hltv": 111,   "esl": 191,     "gosu": 1137},
    "Fluxo":        {"seed": 15,    "hltv": 45,    "esl": 130,     "gosu": 1149},
    "FaZe":         {"seed": 16,    "hltv": 680,   "esl": 1675,    "gosu": 1436},
}

# shape hltv and esl ratings to be more normally distributed
for team in team_data.keys():
    team_data[team]["hltv"] = (team_data[team]["hltv"] ** 0.5) * 10
    team_data[team]["esl"] = (team_data[team]["esl"] ** 0.5) * 10

# empirically tuned to have approx 80% probability of the favourites advancing the tournament
sigma = {
    "hltv": 165,
    "esl": 295,
    "gosu": 425,
}

@cache
def win_probability(first_team, second_team):
    # calculate the win probability of a team with the first rating matched against
    # a team with the second rating given a value of sigma (std deviation of ratings)
    # for each rating system and take the median
    return sorted([1 / (1 + 10 ** ((team_data[second_team][s] - team_data[first_team][s]) / (2 * sigma[s]))) for s in rating_systems])[len(rating_systems) // 2]


class SwissSystem:
    def __init__(self):
        self.clear()

    def clear(self):
        self.teams = {team: {"seed": team_data[team]["seed"], "wins": 0, "losses": 0} for team in team_data.keys()}
        self.finished = dict()

    def simulate_match(self, first_team, second_team):
        # BO3 if match is for advancement/elimination
        is_bo3 = self.teams[first_team]["wins"] == 2 or self.teams[first_team]["losses"] == 2

        # simulate outcome
        probability = win_probability(first_team, second_team)
        if is_bo3:
            first_map = probability > random.random()
            second_map = probability > random.random()

            if first_map != second_map:
                # 1-1 goes to third map
                first_team_win = probability > random.random()
            else:
                # 2-0 no third map
                first_team_win = first_map
        else:
            first_team_win = probability > random.random()

        # update team records
        if first_team_win:
            self.teams[first_team]["wins"] += 1
            self.teams[second_team]["losses"] += 1
        else:
            self.teams[first_team]["losses"] += 1
            self.teams[second_team]["wins"] += 1

        # advance/eliminate teams
        if is_bo3:
            for team in [first_team, second_team]:
                if self.teams[team]["wins"] == 3 or self.teams[team]["losses"] == 3:
                    self.finished[team] = self.teams.pop(team)

    def simulate_round(self):
        # group teams with same record together
        # each group is a list of tuples: (team_seed, team_name)
        even_teams = []
        pos_teams = []
        neg_teams = []

        for team in self.teams.keys():
            if self.teams[team]["wins"] > self.teams[team]["losses"]:
                pos_teams.append((team_data[team]["seed"], team))
            elif self.teams[team]["wins"] < self.teams[team]["losses"]:
                neg_teams.append((team_data[team]["seed"], team))
            else:
                even_teams.append((team_data[team]["seed"], team))

        # sort group by seed and simulate match outcomes
        for group in [even_teams, pos_teams, neg_teams]:
            group.sort(key=lambda x: x[0])

            while group:
                self.simulate_match(group[0][1], group[-1][1])
                group.remove(group[0])
                group.remove(group[-1])

    def simulate_tournament(self):
        # clear data from previous simulation
        if len(self.finished):
            self.clear()

        # simulate whole tournament stage
        while self.teams:
            self.simulate_round()


def simulate_many_tournaments(n):
    # simulate tournament outcomes 'n' times and record statistics
    ss = SwissSystem()
    teams = {team: {stat: 0 for stat in stat_keys} for team in team_data.keys()}

    for i in range(n):
        ss.simulate_tournament()

        for team in ss.finished.keys():
            if ss.finished[team]["wins"] == 3:
                if ss.finished[team]["losses"] == 0:
                    teams[team]["3-0"] += 1
                teams[team]["advance"] += 1
            else:
                if ss.finished[team]["wins"] == 0:
                    teams[team]["0-3"] += 1

    return teams


if __name__ == "__main__":
    # run 'n' simulations total, across 'k' processes
    n = 1_000_000
    k = 16
    teams = {team: {stat: 0 for stat in stat_keys} for team in team_data.keys()}

    start_time = time.time()

    with multiprocessing.Pool(k) as p:
        processes = [p.apply_async(simulate_many_tournaments, [n // k]) for _ in range(k)]
        results = [process.get() for process in processes]

        for result in results:
            for team in teams.keys():
                for stat in stat_keys:
                    teams[team][stat] += result[team][stat]

    # sort and print results
    print(f"RESULTS FROM {n:,} TOURNAMENT SIMULATIONS")
    for stat in stat_keys:
        teams_copy = teams.copy()
        sorted_teams = []

        while teams_copy:
            biggest = {"name": "", "value": 0}

            for team, data in teams_copy.items():
                if data[stat] > biggest["value"]:
                    biggest["value"] = data[stat]
                    biggest["name"] = team

            sorted_teams += [biggest]
            teams_copy.pop(biggest["name"])

        print(f"\nMost likely to {stat}:")

        for i, team in enumerate(sorted_teams):
            print(f"{str(i + 1) + '.' :<3} {team['name'] :<12} {round(team['value'] / n * 100, 2)}%")

    print(f"\nRun time: {round(time.time() - start_time, 3)} seconds")
