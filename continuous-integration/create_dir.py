from argparse import ArgumentParser, Namespace
from pathlib import Path


def _args() -> Namespace:
    parser = ArgumentParser()
    parser.add_argument("--path", type=Path, required=True)
    return parser.parse_args()


def _main():
    path = _args().path
    print(f"Creating directory: {path}")
    path.mkdir(parents=True, exist_ok=True)


if __name__ == "__main__":
    _main()
