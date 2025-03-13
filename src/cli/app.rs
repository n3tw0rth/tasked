use crate::args::{Args, Command, ListOption};
use crate::auth::google::GoogleOAuth;
use crate::common::tasks::GoogleTasks;
use anyhow::Result;

pub struct Cli {
    tasks: GoogleTasks,
}

impl Cli {
    pub async fn new() -> Self {
        Self {
            tasks: GoogleTasks::new().await,
        }
    }

    pub async fn handle_commands(self, args: Args) -> Result<()> {
        match args.command {
            Some(Command::Login) => {
                GoogleOAuth::new().await.sign_in().await?;
            }
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
