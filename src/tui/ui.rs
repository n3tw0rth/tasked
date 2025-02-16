use ratatui::layout::{Constraint, Direction, Layout};
use ratatui::style::{Color, Modifier, Style, Stylize};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Paragraph};
use ratatui::Frame;

use crate::State;

pub enum Tab {
    Todo = 0,
    Calendar = 1,
    Reminders = 2,
}

impl Default for Tab {
    fn default() -> Self {
        Self::Todo
    }
}

impl From<usize> for Tab {
    fn from(v: usize) -> Self {
        match v {
            0 => Self::Todo,
            1 => Self::Calendar,
            2 => Self::Reminders,
            _ => Self::default(),
        }
    }
}

pub fn render(_state: &mut State, frame: &mut Frame) {
    let chunks = Layout::new(
        Direction::Vertical,
        [
            Constraint::Length(3),
            Constraint::Min(10),
            Constraint::Length(3),
        ],
    )
    .vertical_margin(0)
    .split(frame.area());
    {
        let area = chunks[0];
        let lines = vec![
            env!("CARGO_PKG_NAME").fg(Color::Green),
            " ".into(),
            env!("CARGO_PKG_VERSION").fg(Color::Gray).italic(),
        ];

        let paragraph = Paragraph::new(Line::from(lines))
            .centered()
            .block(Block::bordered());
        frame.render_widget(paragraph, area);
    }
    {
        let area = chunks[1];
        let horizontal_chunks = Layout::new(
            Direction::Horizontal,
            [Constraint::Percentage(70), Constraint::Percentage(30)],
        )
        .split(area);
        {
            let area = horizontal_chunks[0];
            frame.render_widget(Block::bordered(), area);
        }
        {
            let area = horizontal_chunks[1];
            frame.render_widget(Block::bordered(), area);
        }
        frame.render_widget(Paragraph::new(""), area);
    }
    {
        let area = chunks[2];
        let key_bindings = vec![
            vec!["", "Navigate"],
            vec!["a", "Add todo"],
            vec!["d", "Delete todo"],
        ];

        let mut line = Line::default();
        key_bindings.iter().for_each(|binding| {
            line.push_span(binding[0].fg(Color::Gray));
            line.push_span(" ");
            line.push_span(binding[1].fg(Color::Green));
            line.push_span(" | ");
        });

        let paragraph = Paragraph::new(line).left_aligned().block(Block::bordered());
        frame.render_widget(paragraph, area);
    }
}
