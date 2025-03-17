use anyhow::Result;
use dotenv::dotenv;

pub fn envs() -> Result<()> {
    dotenv().ok();
    Ok(())
}
