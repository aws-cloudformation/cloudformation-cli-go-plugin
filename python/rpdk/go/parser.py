def setup_subparser(subparsers, parents):
    parser = subparsers.add_parser(
        "go",
        description="This sub command generates IDE and build files for Go",
        parents=parents,
    )
    parser.set_defaults(language="go")

    parser.add_argument(
        "-p",
        "--import-path",
        help="Select the go language import path.",
    )

    return parser
