use ansi_term::Color;
use serde::Deserialize;
use strum_macros::EnumString;

pub mod tasks;

#[derive(Debug, PartialEq, EnumString, Clone, Deserialize)]
pub enum Priority {
    #[strum(serialize = "[p1]")]
    P1,
    P2,
    P3,
    P4,
    P5,
}

impl Default for Priority {
    fn default() -> Self {
        Priority::P5
    }
}

impl Priority {
    pub fn find(str: &str) -> Priority {
        match str {
            "[p1]" => Priority::P1,
            "[p2]" => Priority::P2,
            "[p3]" => Priority::P3,
            "[p4]" => Priority::P4,
            _ => Priority::P5,
        }
    }

    pub fn color(self, str: String) -> String {
        let color = match self {
            Self::P1 => Color::RGB(255, 0, 0),
            Self::P2 => Color::RGB(255, 165, 0),
            Self::P3 => Color::RGB(255, 215, 0),
            Self::P4 => Color::RGB(0, 128, 0),
            Self::P5 => Color::RGB(0, 0, 255),
        };

        color.paint(str).to_string()
    }
}
