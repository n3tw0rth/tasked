use crate::tui::commands::Command;
use anyhow::Result;
use ratatui::crossterm::event::KeyEvent;
#[derive(Default)]
pub struct State {
    pub running: bool,
}

impl State {
    pub fn new() -> Result<Self> {
        let state = Self { running: true };

        Ok(state)
    }
    pub fn find_command(&self, _key_event: KeyEvent) -> Result<Command> {
        Ok(Command::AddTodo)
    }

    pub fn run_command(&self, command: Command) {
        println!("{:?}", command);
    }
}
