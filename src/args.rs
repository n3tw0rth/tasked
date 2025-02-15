#[derive(Parser, Debug)]
#[command(version, about, long_about = None,arg_required_else_help = true,trailing_var_arg=true)]
#[clap(
    about = constants::ABOUT_TEXT,
)]
struct Args {
    #[clap(flatten)]
    global_opts: GlobalOpts,

    #[clap(subcommand)]
    command: Command,
}

#[derive(Debug, Subcommand)]
enum Command {
    /// Help message for read.
    Read {
        /// An example option
        #[clap(long, short = 'o')]
        example_opt: bool,

        /// The path to read from
        path: Utf8PathBuf,
        // (can #[clap(flatten)] other argument structs here)
    },
    /// Help message for write.
    Write(WriteArgs),
    // ...other commands (can #[clap(flatten)] other enum variants here)
}

#[derive(Debug, Args)]
struct GlobalOpts {
    /// Sync
    #[clap(long, arg_enum, global = true, default_value_t = Color::Auto)]
    sync: bool,

    /// Verbosity level (can be specified multiple times)
    #[clap(long, short, global = true, parse(from_occurrences))]
    verbose: usize,
    //... other global options
}
