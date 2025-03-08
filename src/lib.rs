pub mod args;
pub mod auth;
pub mod events;
pub mod prelude;
pub mod tracer;
pub mod tui;

use anyhow::Result;
use ratatui::crossterm::event::KeyCode;
use std::io;

use args::Args;
use events::{Event, EventHandler};
use prelude::{state::State, Tui};

use ratatui::prelude::CrosstermBackend;
use ratatui::Terminal;

pub async fn run(args: Args) -> Result<()> {
    start_tui(args).await
}

pub async fn start_tui(args: Args) -> Result<()> {
    if args.login {
        auth::google::authenticate().await?
    }

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
                if key_event.code == KeyCode::Char('q') {
                    break;
                } else {
                    let command = state.find_command(key_event)?;
                    state.run_command(command);
                }
            }
            Event::Mouse(_mouse_event) => {}
            Event::Resize(_, _) => {}
        }
    }

    // Exit the user interface.
    tui.exit()?;
    Ok(())
}
