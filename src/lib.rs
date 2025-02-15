pub mod events;
pub mod logging;
pub mod prelude;
pub mod tui;

use anyhow::Result;
use ratatui::crossterm::event::KeyCode;
use std::io;

use events::{Event, EventHandler};
use prelude::{state::State, Tui};

use ratatui::prelude::CrosstermBackend;
use ratatui::Terminal;
pub fn run() -> Result<()> {
    start_tui()
}

pub fn start_tui(//args: Args
) -> Result<()> {
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
                    tui.exit()?;
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
