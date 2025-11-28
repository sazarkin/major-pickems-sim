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

    def reset(self) -> None:
        """Reset state for new simulation."""
        for record in self.records.values():
            record.wins = 0
            record.losses = 0
        for team_set in self.faced.values():
            team_set.clear()
        self.remaining.clear()
        self.remaining.update(self.records.keys())
        self.finished.clear()

    def simulate_match(self, team_a: Team, team_b: Team) -> None:
        """Simulate singular match."""
        # BO3 if match is for advancement/elimination
        is_bo3 = self.records[team_a].wins == 2 or self.records[team_a].losses == 2

        # calculate single map win probability
        p = win_probability(team_a, team_b, self.sigma)

        # simulate match outcome
        if is_bo3:
            # Simulate proper BO3 series where first to 2 wins takes the match
            a_wins, b_wins = 0, 0
            while a_wins < 2 and b_wins < 2:
                if p > random.random():
                    a_wins += 1
                else:
                    b_wins += 1
            team_a_win = a_wins > b_wins
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
    ) -> Tuple[dict[Team, dict[str, int]], list[Tuple[int, int]]]:
        """Run batch of 'n' simulation iterations for given data and return results."""
        results = {
            team: {stat: 0 for stat in ["3-0", "3-1 or 3-2", "0-3"]}
            for team in self.teams
        }
        success_counts = [0] * len(predictions)

        # Pre-convert prediction team lists to sets for faster intersection
        prediction_sets = [
            {
                "3-0": set(p["3-0"]),
                "3-1 or 3-2": set(p["3-1 or 3-2"]),
                "0-3": set(p["0-3"]),
            }
            for p in predictions
        ]

        # Instantiate SwissSystem once, outside the loop
        ss = SwissSystem(
            self.sigma,
            {team: Record(0, 0) for team in self.teams},
            {team: set() for team in self.teams},
            set(self.teams),
            set(),
        )

        for _ in range(n):
            ss.reset()  # Reset state instead of re-creating
            ss.simulate_tournament()

            outcome_groups = defaultdict(set)
            for team, record in ss.records.items():
                if record.wins == 3:
                    if record.losses == 0:
                        results[team]["3-0"] += 1
                        outcome_groups["3-0"].add(team.name)
                    else:
                        results[team]["3-1 or 3-2"] += 1
                        outcome_groups["3-1 or 3-2"].add(team.name)
                elif record.losses == 3:
                    results[team]["0-3"] += 1
                    outcome_groups["0-3"].add(team.name)

            # Optimized scoring using set intersections
            for i, p_sets in enumerate(prediction_sets):
                score = len(outcome_groups["3-0"].intersection(p_sets["3-0"]))
                score += len(
                    outcome_groups["3-1 or 3-2"].intersection(p_sets["3-1 or 3-2"])
                )
                score += len(outcome_groups["0-3"].intersection(p_sets["0-3"]))
                if score >= 6:
                    success_counts[i] += 1

        return results, list(zip(success_counts, [n] * len(predictions)))

    def run(
        self, n: int, k: int, predictions
    ) -> Tuple[dict[Team, dict[str, int]], list[float]]:
        """Run 'n' simulation iterations across 'k' processes and return results."""
        batch_size = n // k
        remainder = n % k

        with Pool(k) as pool:
            futures = [
                pool.apply_async(
                    self.batch, [batch_size + (1 if i < remainder else 0), predictions]
                )
                for i in range(k)
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

        # Aggregate success counts and simulation counts per prediction
        combined_success_counts = [0] * len(predictions)
        combined_simulation_counts = [0] * len(predictions)

        for _, batch_stats in results:
            for i, (success_count, sim_count) in enumerate(batch_stats):
                combined_success_counts[i] += success_count
                combined_simulation_counts[i] += sim_count

        percentages = [
            (
                (combined_success_counts[i] / combined_simulation_counts[i]) * 100
                if combined_simulation_counts[i] > 0
                else 0.0
            )
            for i in range(len(predictions))
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
    prediction_teams = []
    prediction_teams.extend(prediction["3-0"])
    prediction_teams.extend(prediction["3-1 or 3-2"])
    prediction_teams.extend(prediction["0-3"])

    # Ensure we include all teams exactly once
    prediction_teams += [team for team in teams if team not in prediction_teams]

    random_group = random.choice(["3-0", "3-1 or 3-2", "0-3"])
    random_team_a = random.choice(prediction[random_group])
    random_team_b = random.choice(
        [t for t in prediction_teams if t not in prediction[random_group]]
    )

    index_a = prediction_teams.index(random_team_a)
    index_b = prediction_teams.index(random_team_b)
    prediction_teams[index_a], prediction_teams[index_b] = (
        prediction_teams[index_b],
        prediction_teams[index_a],
    )

    return {
        "3-0": prediction_teams[:2],
        "3-1 or 3-2": prediction_teams[2:8],
        "0-3": prediction_teams[8:10],
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
    predictions = []
    # Generate predictions first then make copy of hashes
    prediction_hashes = set()
    prediction_hashes_copy = (
        set()
    )  # Initialize empty, will fill after generating predictions

    # Generate unique random predictions
    for _ in range(args.p):
        while True:
            # Create fresh random prediction
            shuffled = random.sample(teams, len(teams))
            prediction = {
                "3-0": shuffled[:2],
                "3-1 or 3-2": shuffled[2:8],
                "0-3": shuffled[8:10],
            }
            ph = hash_prediction(prediction)
            if ph not in prediction_hashes:
                prediction_hashes.add(ph)
                predictions.append(prediction)
                break

    start = perf_counter_ns()
    results, scores = Simulation(args.f).run(args.n, args.k, predictions)
    run_time = (perf_counter_ns() - start) / 1_000_000_000

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
