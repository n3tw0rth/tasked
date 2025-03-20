use anyhow::Result;
use ratatui::prelude::Backend;
use ratatui::text::Line;
use ratatui::widgets::Block;
use ratatui::{Frame, Terminal, TerminalOptions, Viewport};

pub struct Inline {}

impl Inline {
    pub fn new() -> Self {
        Self {}
    }

    pub fn show<G>(self, f: G) -> Result<()>
    where
        G: FnOnce(&mut Frame),
    {
        color_eyre::install().unwrap();
        let mut terminal = ratatui::init_with_options(TerminalOptions {
            viewport: Viewport::Inline(8),
        });

        let _app_result = self.run(&mut terminal, f);

        ratatui::restore();

        Ok(())
    }

    fn run<G>(self, terminal: &mut Terminal<impl Backend>, f: G) -> Result<()>
    where
        G: FnOnce(&mut Frame),
    {
        terminal.draw(|frame| self.draw(frame, f))?;
        Ok(())
    }

    fn draw<G>(&self, frame: &mut Frame, f: G)
    where
        G: FnOnce(&mut Frame),
    {
        f(frame);
    }
}
