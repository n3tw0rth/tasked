use clap::{Parser, Subcommand};

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
    #[clap(subcommand)]
    pub command: Option<Command>,
}

#[derive(Clone, Debug, Subcommand)]
pub enum Command {
    /// Login to Google using OAuth
    Login,

    /// Set List
    #[clap(name = "list")]
    List { list: ListOption },

    /// Add something
    Add { value: String },
}

#[derive(Debug, Clone, Copy, clap::ValueEnum)]
pub enum ListOption {
    /// Tasks Lists
    Lists,
    /// Tasks
    Tasks,
}
