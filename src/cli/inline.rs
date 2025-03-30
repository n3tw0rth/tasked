use anyhow::Result;

pub struct Inline {}

impl Inline {
    pub fn new() -> Self {
        Self {}
    }

    pub fn show<G>(self, f: G) -> Result<()>
    where
        G: FnOnce(),
    {
        let _app_result = self.run(f);

        Ok(())
    }

    fn run<G>(self, f: G) -> Result<()>
    where
        G: FnOnce(),
    {
        self.draw(f);
        Ok(())
    }

    fn draw<G>(&self, f: G)
    where
        G: FnOnce(),
    {
        f();
    }
}
