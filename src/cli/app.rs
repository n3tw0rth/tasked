use crate::args::{Args, AuthOption, Command, ListOption};
use crate::auth::google::GoogleOAuth;
use crate::common::tasks::GoogleTasks;

use anyhow::Result;
use retry::retry;

pub struct Cli {
    tasks: GoogleTasks,
}

impl Cli {
    pub async fn new() -> Self {
        Self {
            tasks: GoogleTasks::new().await,
        }
    }

    pub async fn handle_commands(mut self, args: Args) -> Result<()> {
        let mut google_auth = GoogleOAuth::new().await;
        match args.command {
            Some(Command::Auth { auth }) => match auth {
                AuthOption::Login => google_auth.sign_in().await?,
                AuthOption::Refresh => google_auth.refresh_token().await?,
                AuthOption::Logout => unimplemented!(),
            },
            Some(Command::List { list }) => match list {
                ListOption::Lists => self.tasks.get_tasks_lists().await?,
                ListOption::Tasks => self.tasks.get_tasks(&"").await?,
            },
            Some(Command::Add { value }) => {
                println!("Adding {}", value);
                self.tasks.add_tasks().await?;
            }
            None => {
                println!("No command provided.");
            }
        }

        Ok(())
    }
}
