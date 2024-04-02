import sys
from argparse import ArgumentParser, Namespace


def _args() -> Namespace:
    parser = ArgumentParser()
    parser.add_argument("--push-status", type=str, required=True)
    return parser.parse_args()


def _main():
    push_status = _args().push_status
    print("Checking push status")
    if push_status == "true":
        print("Commit has been pushed successfully")
        sys.exit(0)
    else:
        print("Commit has not been pushed successfully")
        sys.exit(1)


if __name__ == "__main__":
    _main()
