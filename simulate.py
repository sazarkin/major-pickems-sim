from __future__ import annotations

import json
import random
import hashlib
from argparse import ArgumentParser
from dataclasses import dataclass
from functools import cache, reduce
from multiprocessing import Pool
from os import cpu_count
from statistics import median
from time import perf_counter_ns
from typing import TYPE_CHECKING, Tuple
from collections import defaultdict

if TYPE_CHECKING:
    from pathlib import Path


@dataclass
class Team:
    name: str
    seed: int
    rating: tuple[int, ...]

    def __repr__(self) -> str:
        return str(self.name)

    def __hash__(self) -> int:
        return hash(self.name)

    def __eq__(self, other: Team) -> bool:  # type: ignore
        return self.name == other.name


@dataclass
class Record:
    wins: int
    losses: int

    @property
    def diff(self) -> int:
        return self.wins - self.losses


@cache
def win_probability(a: Team, b: Team, sigma: tuple[int, ...]) -> float:
    """Calculate the probability of team 'a' beating team 'b' for given sigma values."""
    # calculate the win probability for given team ratings and value of sigma (std deviation of
    # ratings) for each rating system (assumed to be elo based and normally distributed) and
    # take the median
    return median(
        1 / (1 + 10 ** ((b.rating[i] - a.rating[i]) / (2 * sigma[i])))
        for i in range(len(sigma))
    )


@dataclass
class SwissSystem:
    sigma: tuple[int, ...]
    records: dict[Team, Record]
    faced: dict[Team, set[Team]]
    remaining: set[Team]
    finished: set[Team]

    def seeding(self, team: Team) -> tuple[int, int, int]:
        """Calculate seeding based on win-loss, Buchholz difficulty, and initial seed."""
        return (
            -self.records[team].diff,
            -sum(self.records[opp].diff for opp in self.faced[team]),
            team.seed,
        )

    def simulate_match(self, team_a: Team, team_b: Team) -> None:
        """Simulate singular match."""
        # BO3 if match is for advancement/elimination
        is_bo3 = self.records[team_a].wins == 2 or self.records[team_a].losses == 2

        # calculate single map win probability
        p = win_probability(team_a, team_b, self.sigma)

        # simulate match outcome
        if is_bo3:
            first_map = p > random.random()
            second_map = p > random.random()
            team_a_win = p > random.random() if first_map != second_map else first_map
        else:
            team_a_win = p > random.random()

        # update team records
        if team_a_win:
            self.records[team_a].wins += 1
            self.records[team_b].losses += 1
        else:
            self.records[team_a].losses += 1
            self.records[team_b].wins += 1

        # add to faced teams
        self.faced[team_a].add(team_b)
        self.faced[team_b].add(team_a)

        # advance/eliminate teams after best of three
        if is_bo3:
            for team in [team_a, team_b]:
                if self.records[team].wins == 3 or self.records[team].losses == 3:
                    self.remaining.remove(team)
                    self.finished.add(team)

    def simulate_round(self) -> None:
        """Simulate round of matches."""
        even_teams, pos_teams, neg_teams = [], [], []

        # group teams with the same record together and sort by mid-round seeding
        for team in sorted(self.remaining, key=self.seeding):
            if self.records[team].diff > 0:
                pos_teams.append(team)
            elif self.records[team].diff < 0:
                neg_teams.append(team)
            else:
                even_teams.append(team)

        # first round is seeded differently (1-9, 2-10, 3-11 etc.)
        if len(even_teams) == len(self.records):
            half = len(even_teams) // 2
            even_teams[half:] = reversed(even_teams[half:])

        # run matches for each group, highest seed vs lowest seed
        for group in [pos_teams, even_teams, neg_teams]:
            half = len(group) // 2

            for a, b in zip(group[:half], reversed(group[half:])):
                self.simulate_match(a, b)

    def simulate_tournament(self) -> None:
        """Simulate entire tournament stage."""
        while self.remaining:
            self.simulate_round()


class Simulation:
    sigma: tuple[int, ...]
    teams: set[Team]

    def __init__(self, filepath: Path) -> None:
        """Parse data loaded in from .json file."""
        with open(filepath) as file:
            data = json.load(file)

        self.sigma = (*data["sigma"].values(),)
        self.teams = {
            Team(
                team_k,
                team_v["seed"],
                tuple(
                    (eval(sys_v))(team_v[sys_k])
                    for sys_k, sys_v in data["systems"].items()
                ),  # noqa: S307
            )
            for team_k, team_v in data["teams"].items()
        }

    def batch(
        self, n: int, predictions: list[dict]
    ) -> Tuple[dict[Team, dict[str, int]], list[int]]:
        """Run batch of 'n' simulation iterations for given data and return results."""
        results = {
            team: {stat: 0 for stat in ["3-0", "3-1 or 3-2", "0-3"]}
            for team in self.teams
        }
        scores = [[] for _ in predictions]

        for _ in range(n):
            ss = SwissSystem(
                self.sigma,
                {team: Record(0, 0) for team in self.teams},
                {team: set() for team in self.teams},
                set(self.teams),
                set(),
            )

            ss.simulate_tournament()

            for team, record in ss.records.items():
                if record.wins == 3:
                    if record.losses == 0:
                        results[team]["3-0"] += 1
                    else:
                        results[team]["3-1 or 3-2"] += 1
                elif record.wins == 0:
                    results[team]["0-3"] += 1

            for i, prediction in enumerate(predictions):
                score = 0
                for team, record in ss.records.items():
                    key = f"{record.wins}-{record.losses}"
                    if key == "3-1" or key == "3-2":
                        key = "3-1 or 3-2"
                    key_teams = prediction.get(key)
                    if key_teams and team.name in key_teams:
                        score += 1
                scores[i].append(score)

        return results, scores

    def run(
        self, n: int, k: int, predictions
    ) -> Tuple[dict[Team, dict[str, int]], int]:
        """Run 'n' simulation iterations across 'k' processes and return results."""
        with Pool(k) as pool:
            futures = [
                pool.apply_async(self.batch, [n // k, predictions]) for _ in range(k)
            ]
            results = [future.get() for future in futures]

        def _f(acc: dict, res: dict) -> dict:
            for team, result in res.items():
                for key, val in result.items():
                    acc[team][key] += val
            return acc

        combined_results = reduce(
            _f, map(lambda x: x[0], results), defaultdict(lambda: defaultdict(int))
        )

        scores = [[] for _ in predictions]
        for _, batch_scores in results:
            for i, score in enumerate(batch_scores):
                scores[i].extend(score)

        percentages = [
            (sum(1 for score in score_list if score > 5) / len(score_list)) * 100
            for score_list in scores
        ]

        return combined_results, percentages


def format_results(
    results: dict[Team, dict[str, int]], n: int, run_time: float
) -> list[str]:
    """Formats simulation results and run time parameters into readable string."""
    out = [f"RESULTS FROM {n:,} TOURNAMENT SIMULATIONS"]

    for stat in next(iter(results.values())):
        out.append(f"\nMost likely to {stat}:")

        for i, (team, result) in enumerate(
            sorted(results.items(), key=lambda tup: tup[1][stat], reverse=True),
        ):
            out.append(
                f"{str(i + 1) + '.':<3} {team.name:<15} {round(result[stat] / n * 100, 1):>5}%",
            )

    out.append(f"\nRun time: {run_time:.2f} seconds")
    return out


def mutate_prediction(prediction: dict, teams: list) -> dict:
    predition_teams = []
    predition_teams.extend(prediction["3-0"])
    predition_teams.extend(prediction["3-1 or 3-2"])
    predition_teams.extend(prediction["0-3"])

    for team in teams:
        if team not in predition_teams:
            predition_teams.append(team)

    random_group = random.choice(["3-0", "3-1 or 3-2", "0-3"])
    random_team_a = random.choice(prediction[random_group])
    random_team_b = random.choice(
        [t for t in predition_teams if t not in prediction[random_group]]
    )

    index_a = predition_teams.index(random_team_a)
    index_b = predition_teams.index(random_team_b)
    predition_teams[index_a], predition_teams[index_b] = (
        predition_teams[index_b],
        predition_teams[index_a],
    )

    return {
        "3-0": predition_teams[:2],
        "3-1 or 3-2": predition_teams[2:8],
        "0-3": predition_teams[8:10],
    }


def hash_prediction(prediction: dict) -> str:
    a = (
        sorted(prediction["3-0"])
        + sorted(prediction["3-1 or 3-2"])
        + sorted(prediction["0-3"])
    )
    return hashlib.md5(str(a).encode("utf-8")).hexdigest()


if __name__ == "__main__":
    # parse args from CLI
    parser = ArgumentParser()
    parser.add_argument(
        "-f", type=str, help="path to input data (.json)", required=True
    )
    parser.add_argument(
        "-n", type=int, default=1_000_000, help="number of iterations to run"
    )
    parser.add_argument(
        "-k", type=int, default=cpu_count(), help="number of cores to use"
    )
    parser.add_argument("-p", type=int, default=1, help="number of predictions to run")
    parser.add_argument("-s", type=int, default=0, help="random seed")
    args = parser.parse_args()

    if args.s:
        random.seed(args.s)

    data = json.load(open(args.f))
    teams = list(data["teams"].keys())
    predictions = [
        {
            "3-0": ["G2", "Vitality"],
            "3-1 or 3-2": [
                "The MongolZ",
                "HEROIC",
                "Spirit",
                "MOUZ",
                "FaZe",
                "Natus Vincere",
            ],
            "0-3": ["GamerLegion", "MIBR"],
        },
        {
            "3-0": ["G2", "Vitality"],
            "3-1 or 3-2": [
                "The MongolZ",
                "HEROIC",
                "Spirit",
                "MOUZ",
                "FaZe",
                "Natus Vincere",
            ],
            "0-3": ["MIBR", "Wildcard"],
        },
        {
            "3-0": ["G2", "Natus Vincere"],
            "3-1 or 3-2": [
                "The MongolZ",
                "HEROIC",
                "Spirit",
                "MOUZ",
                "FaZe",
                "Vitality",
            ],
            "0-3": ["GamerLegion", "MIBR"],
        },
        {
            "3-0": ["G2", "Natus Vincere"],
            "3-1 or 3-2": [
                "The MongolZ",
                "HEROIC",
                "Spirit",
                "MOUZ",
                "FaZe",
                "Vitality",
            ],
            "0-3": ["MIBR", "Wildcard"],
        },
        {
            "3-0": ["Natus Vincere", "Vitality"],
            "3-1 or 3-2": ["G2", "HEROIC", "FaZe", "MOUZ", "The MongolZ", "Spirit"],
            "0-3": ["MIBR", "GamerLegion"],
        },
    ]
    prediction_hashes = set([hash_prediction(p) for p in predictions])
    prediction_hashes_copy = prediction_hashes.copy()

    for _ in range(args.p):
        base_prediction = random.choice(predictions)
        for i in range(10):
            prediction = base_prediction.copy()
            for _ in range(random.randint(1, 5)):
                prediction = mutate_prediction(prediction, teams)
            prediction_hash = hash_prediction(prediction)
            if prediction_hash not in prediction_hashes:
                prediction_hashes.add(prediction_hash)
                predictions.append(prediction)
                break

    start = perf_counter_ns()
    results, scores = Simulation(args.f).run(args.n, args.k, predictions)
    run_time = (perf_counter_ns() - start) / 1_000_000_000
    # print("\n".join(format_results(results, args.n, run_time)))

    prediction_results = list(zip(scores, predictions))
    prediction_results.sort(key=lambda x: x[0], reverse=True)
    for score, prediction in prediction_results[:5]:
        print(f"Percent of success: {score:.2f}%")
        h = hash_prediction(prediction)
        if h in prediction_hashes_copy:
            print(f"\033[1m{h[-5:]}\033[0m")
        else:
            print(f"{h[-5:]}")
        for key, value in prediction.items():
            print(f"'{key}': {value}")
        print()
