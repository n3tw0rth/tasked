pub mod args;
pub mod auth;
pub mod cli;
pub mod cmd;
pub mod common;
pub mod events;
pub mod prelude;
pub mod tracer;
pub mod tui;
pub mod utils;

use anyhow::Result;
use std::io;

use args::Args;
use events::{Event, EventHandler};
use prelude::{state::State, Tui};

use ratatui::prelude::CrosstermBackend;
use ratatui::Terminal;

use cli::app::Cli;
use cmd::commands::Command;

pub async fn run(args: Args) -> Result<()> {
    utils::envs()?;
    // Check if no arguments are passed
    if std::env::args().len() > 1 {
        // If there are arguments passed run CLI
        start_cli(args).await
    } else {
        // If there are no arguments run TUI
        start_tui().await
    }
}

pub async fn start_cli(args: Args) -> Result<()> {
    let mut cli = Cli::new().await;
    cli.handle_commands(args).await?;

    Ok(())
}

pub async fn start_tui() -> Result<()> {
    // Create an application.
    let mut state = State::new()?;

    // Initialize the terminal user interface.
    let backend = CrosstermBackend::new(io::stdout());
    let terminal = Terminal::new(backend)?;
    let events = EventHandler::new(250);
    let mut tui = Tui::new(terminal, events);
    tui.init()?;

    // Start the main loop.
    while state.running {
        // Render the user interface.
        tui.draw(&mut state)?;
        // Handle events.
        match tui.events.next()? {
            Event::Tick => {}
            Event::Key(key_event) => {
                let command = Command::from(key_event);
                state.run_command(command, tui.events.sender.clone())?;
            }
            Event::Mouse(_mouse_event) => {}
            Event::Resize(_, _) => {}
        }
    }

    // Exit the user interface.
    tui.exit()?;
    Ok(())
}
