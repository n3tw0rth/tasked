use ratatui::widgets::Paragraph;
use ratatui::Frame;

use crate::State;

pub fn render(_state: &mut State, frame: &mut Frame) {
    let paragraph = Paragraph::new("hello");

    let area = frame.area();
    frame.render_widget(paragraph, area);
}
