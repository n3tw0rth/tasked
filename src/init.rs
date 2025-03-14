use anyhow::Result;
use dotenv::dotenv;

pub fn init() -> Result<()> {
    dotenv().ok();
    Ok(())
}
