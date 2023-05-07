import random
import multiprocessing
import time
import functools
from pprint import pprint


rating_systems = ["hltv", "esl", "gosu"]

team_ratings = {
    "Monte":        {"hltv": 113,   "esl": 182,     "gosu": 1218},
    "paiN":         {"hltv": 178,   "esl": 442,     "gosu": 1232},
    "G2":           {"hltv": 697,   "esl": 1322,    "gosu": 1553},
    "GamerLegion":  {"hltv": 78,    "esl": 107,     "gosu": 1184},
    "FORZE":        {"hltv": 195,   "esl": 419,     "gosu": 1240},
    "Apeks":        {"hltv": 75,    "esl": 80,      "gosu": 1185},
    "NiP":          {"hltv": 216,   "esl": 350,     "gosu": 1262},
    "OG":           {"hltv": 239,   "esl": 292,     "gosu": 1293},
    "ENCE":         {"hltv": 290,   "esl": 559,     "gosu": 1313},
    "MOUZ":         {"hltv": 239,   "esl": 409,     "gosu": 1256},
    "Liquid":       {"hltv": 418,   "esl": 634,     "gosu": 1358},
    "Grayhound":    {"hltv": 101,   "esl": 95,      "gosu": 1066},
    "Complexity":   {"hltv": 161,   "esl": 301,     "gosu": 1158},
    "TheMongolz":   {"hltv": 111,   "esl": 191,     "gosu": 1137},
    "Fluxo":        {"hltv": 45,    "esl": 130,     "gosu": 1149},
    "FaZe":         {"hltv": 680,   "esl": 1675,    "gosu": 1436},
}

# shape hltv and esl ratings to be more normally distributed
for team in team_ratings.keys():
    team_ratings[team]["hltv"] = (team_ratings[team]["hltv"] ** 0.5) * 10
    team_ratings[team]["esl"] = (team_ratings[team]["esl"] ** 0.5) * 10

# empirically tuned to have approx 80% probability of the favourites advancing the tournament
sigma = {
    "hltv": 165,
    "esl": 295,
    "gosu": 425,
}

@functools.cache
def win_probability(first_team, second_team):
    # calculate the win probability of a team with the first rating matched against
    # a team with the second rating given a value of sigma (std deviation of ratings)
    # for each rating system and take the mean
    return sum(1 / (1 + 10 ** ((team_ratings[second_team][s] - team_ratings[first_team][s]) / (2 * sigma[s]))) for s in rating_systems) / len(rating_systems)


class SwissSystem:
    def __init__(self):
        self.clear()

    def clear(self):
        self.finished = dict()
        self.teams = {
            "Monte":        {"seed": 1,     "wins": 0,  "losses": 0},
            "paiN":         {"seed": 2,     "wins": 0,  "losses": 0},
            "G2":           {"seed": 3,     "wins": 0,  "losses": 0},
            "GamerLegion":  {"seed": 4,     "wins": 0,  "losses": 0},
            "FORZE":        {"seed": 5,     "wins": 0,  "losses": 0},
            "Apeks":        {"seed": 6,     "wins": 0,  "losses": 0},
            "NiP":          {"seed": 7,     "wins": 0,  "losses": 0},
            "OG":           {"seed": 8,     "wins": 0,  "losses": 0},
            "ENCE":         {"seed": 9,     "wins": 0,  "losses": 0},
            "MOUZ":         {"seed": 10,    "wins": 0,  "losses": 0},
            "Liquid":       {"seed": 11,    "wins": 0,  "losses": 0},
            "Grayhound":    {"seed": 12,    "wins": 0,  "losses": 0},
            "Complexity":   {"seed": 13,    "wins": 0,  "losses": 0},
            "TheMongolz":   {"seed": 14,    "wins": 0,  "losses": 0},
            "Fluxo":        {"seed": 15,    "wins": 0,  "losses": 0},
            "FaZe":         {"seed": 16,    "wins": 0,  "losses": 0},
        }

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
        even_teams = []
        pos_teams = []
        neg_teams = []

        for team in self.teams.keys():
            if self.teams[team]["wins"] > self.teams[team]["losses"]:
                pos_teams += [team]
            elif self.teams[team]["wins"] < self.teams[team]["losses"]:
                neg_teams += [team]
            else:
                even_teams += [team]

        # match up teams within each group according to seed
        for group in [even_teams, pos_teams, neg_teams]:
            while group:
                highest_seed = group[0]
                lowest_seed = group[-1]

                for team in group:
                    if self.teams[team]["seed"] > self.teams[highest_seed]["seed"]:
                        highest_seed = team
                    if self.teams[team]["seed"] < self.teams[lowest_seed]["seed"]:
                        lowest_seed = team

                group.remove(highest_seed)
                group.remove(lowest_seed)

                # simulate match outcome
                self.simulate_match(highest_seed, lowest_seed)

    def simulate_tournament(self):
        # simulate whole tournament stage
        self.clear()
        while self.teams:
            self.simulate_round()


def simulate_many_tournaments(n):
    # simulate tournament outcomes 'n' times and record statistics
    teams = dict()
    ss = SwissSystem()

    for team_name in team_ratings.keys():
        teams[team_name] = {
            "advance": 0,
            "3-0": 0,
            "0-3": 0
        }

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
    teams = dict()

    for team_name in team_ratings.keys():
        teams[team_name] = {
            "advance": 0,
            "3-0": 0,
            "0-3": 0
        }

    start_time = time.time()

    with multiprocessing.Pool(k) as p:
        processes = [p.apply_async(simulate_many_tournaments, [n // k]) for _ in range(k)]
        results = [process.get() for process in processes]

        for result in results:
            for team in teams.keys():
                for stat in ["advance", "3-0", "0-3"]:
                    teams[team][stat] += result[team][stat]

    # sort and print results
    print(f"RESULTS FROM {n:,} TOURNAMENT SIMULATIONS")
    for stat in ["advance", "3-0", "0-3"]:
        teams_copy = teams.copy()
        sorted_teams = []

        while teams_copy:
            biggest = {
                "name": "",
                "value": 0
            }

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
