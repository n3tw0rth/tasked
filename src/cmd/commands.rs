use ratatui::crossterm::event::{KeyCode, KeyEvent};

/// Possible scroll areas.
#[derive(Debug, PartialEq, Eq)]
pub enum ScrollType {
    /// Main application tabs.
    Tab,
    /// Inner tables.
    Table,
    /// Main list.
    List,
    /// Block.
    Block,
}

#[derive(Debug)]
pub enum Command {
    Next(ScrollType, u8),
    Nothing,
    Add,
    Delete,
    Update,
    Exit,
}

impl From<KeyEvent> for Command {
    fn from(key_event: KeyEvent) -> Self {
        match key_event.code {
            KeyCode::Right | KeyCode::Char('l') => Self::Next(ScrollType::Table, 1),
            KeyCode::Char('q') => Self::Exit,
            _ => Self::Nothing,
        }
    }
}
