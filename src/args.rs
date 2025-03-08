use clap::Parser;

/// Argument parser powered by [`clap`].
#[derive(Clone, Debug, Default, Parser)]
#[clap(
    version,
    author = clap::crate_authors!("\n"),
    about,
    rename_all_env = "screaming-snake",
    help_template = "\
{before-help}{name} {version}
{author-with-newline}{about-with-newline}
{usage-heading}
  {usage}

{all-args}{after-help}
",
)]
pub struct Args {
    /// Login to google using oauth
    #[arg(long)]
    pub login: bool,

    #[arg(short = 'p', long = "project", default_value = "general")]
    pub project: String,
    //
    #[arg(long)]
    pub add: Option<String>,
    //
    //lables: Option<Vec<String>>,
}
