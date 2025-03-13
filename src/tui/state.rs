use std::sync::mpsc;

use crate::cmd::commands::Command;
use crate::events::Event;
use anyhow::Result;
use ratatui::crossterm::event::KeyEvent;
use tracing::info;

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
        Ok(Command::Add)
    }

    pub fn run_command(
        &mut self,
        command: Command,
        _event_sender: mpsc::Sender<Event>,
    ) -> Result<()> {
        info!("{:?}", command);

        match command {
            Command::Exit => {
                self.running = false;
            }
            _ => {}
        }

        Ok(())
    }
}
