import random
import multiprocessing
import time
import functools


rating_systems = ["hltv", "esl", "gosu"]

team_ratings = {
    "FaZe":         {"hltv": 819,   "esl": 1552,    "gosu": 1511},
    "Vitality":     {"hltv": 756,   "esl": 1510,    "gosu": 1460},
    "Liquid":       {"hltv": 741,   "esl": 1374,    "gosu": 1358},
    "NaVi":         {"hltv": 660,   "esl": 1302,    "gosu": 1556},
    "Cloud9":       {"hltv": 595,   "esl": 824,     "gosu": 1275},
    "Heroic":       {"hltv": 476,   "esl": 536,     "gosu": 1382},
    "Outsiders":    {"hltv": 431,   "esl": 476,     "gosu": 1332},
    "MOUZ":         {"hltv": 403,   "esl": 506,     "gosu": 1352},
    "FURIA":        {"hltv": 381,   "esl": 454,     "gosu": 1296},
    "NiP":          {"hltv": 336,   "esl": 539,     "gosu": 1379},
    "ENCE":         {"hltv": 322,   "esl": 434,     "gosu": 1305},
    "Spirit":       {"hltv": 322,   "esl": 395,     "gosu": 1332},
    "BIG":          {"hltv": 204,   "esl": 296,     "gosu": 1296},
    "fnatic":       {"hltv": 156,   "esl": 223,     "gosu": 1254},
    "BNE":          {"hltv": 122,   "esl": 105,     "gosu": 1186},
    "Sprout":       {"hltv": 92,    "esl": 169,     "gosu": 1152},
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

def Q(first_rating, second_rating, sigma):
    # calculate the expected result of a team with the first rating matched against
    # a team with the second rating given a value of sigma (std deviation of ratings)
    # 0 == draw, >0 == first team wins, <0 == second team wins
    return (second_rating - first_rating) / (sigma * 2)

def memoize(function):
    # caches function results to return instead of rerunning expensive calculation
    cache = {}

    @functools.wraps(function)
    def wrapper(*args):
        key = str(args)

        if key not in cache:            
            cache[key] = function(*args)

        return cache[key]

    return wrapper

@memoize
def win_probability(first_team, second_team):
    # calculate Q across all rating systems and take the mean, use that to calculate win probability
    return 1 / (1 + 10 ** sum(Q(team_ratings[first_team][s], team_ratings[second_team][s], sigma[s]) for s in rating_systems) / len(rating_systems))


class SwissSystem:
    def __init__(self):
        self.clear()

    def clear(self):
        self.finished = dict()
        self.teams = {
            "FaZe":         {"seed": 1,     "wins": 0,  "losses": 0},
            "NaVi":         {"seed": 2,     "wins": 0,  "losses": 0},
            "NiP":          {"seed": 3,     "wins": 0,  "losses": 0},
            "ENCE":         {"seed": 4,     "wins": 0,  "losses": 0},
            "Sprout":       {"seed": 5,     "wins": 0,  "losses": 0},
            "Heroic":       {"seed": 6,     "wins": 0,  "losses": 0},
            "Spirit":       {"seed": 7,     "wins": 0,  "losses": 0},
            "Liquid":       {"seed": 8,     "wins": 0,  "losses": 0},
            "MOUZ":         {"seed": 9,     "wins": 0,  "losses": 0},
            "BNE":          {"seed": 10,    "wins": 0,  "losses": 0},
            "Outsiders":    {"seed": 11,    "wins": 0,  "losses": 0},
            "BIG":          {"seed": 12,    "wins": 0,  "losses": 0},
            "FURIA":        {"seed": 13,    "wins": 0,  "losses": 0},
            "fnatic":       {"seed": 14,    "wins": 0,  "losses": 0},
            "Vitality":     {"seed": 15,    "wins": 0,  "losses": 0},
            "Cloud9":       {"seed": 16,    "wins": 0,  "losses": 0},
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
    n = 100000
    k = 1
    teams = dict()
    start_time = time.time()

    with multiprocessing.Pool() as p:
        results = p.map(simulate_many_tournaments, [n // k] * k)
    
    # concatenate results from all processes
    for team in results[0].keys():
        teams[team] = {key: sum(result[team][key] for result in results) for key in results[0][team].keys()}

    # sort and print results to console
    # print(f"RESULTS FROM {n:,} TOURNAMENT SIMULATIONS")
    # for stat in ["advance", "3-0", "0-3"]:
    #     teams_copy = teams.copy()
    #     sorted_teams = []

    #     while teams_copy:
    #         biggest = {
    #             "name": "",
    #             "value": 0
    #         }

    #         for team, data in teams_copy.items():
    #             if data[stat] > biggest["value"]:
    #                 biggest["value"] = data[stat]
    #                 biggest["name"] = team
            
    #         sorted_teams += [biggest]
    #         teams_copy.pop(biggest["name"])

    #     print(f"\nMost likely to {stat}:")
        
    #     for i, team in enumerate(sorted_teams):
    #         print(f"{str(i + 1) + '.' :<3} {team['name'] :<12} {round(team['value'] / n * 100, 2)}%")

    print(f"\nRun time: {round(time.time() - start_time, 3)} seconds")
