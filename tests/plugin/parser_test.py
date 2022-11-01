import argparse

from rpdk.go.parser import setup_subparser


def test_setup_subparser():
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="subparser_name")

    sub_parser = setup_subparser(subparsers, [])

    args = sub_parser.parse_args(["-p", "/path/"])

    assert args.language == "go"
    assert args.import_path == "/path/"
