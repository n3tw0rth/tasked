use anyhow::Result;
use tracing::info;

use crate::auth::google::GoogleOAuth;

const TASKS_API_ENDPOINT: &str = "https://www.googleapis.com/discovery/v1/apis/tasks/v1/";

pub struct GoogleTasks {
    pub access_token: String,
}

impl GoogleTasks {
    pub async fn new() -> Self {
        Self {
            access_token: GoogleOAuth::new().get_tokens().await.unwrap(),
        }
    }

    pub async fn get_tasks_lists(self) -> Result<()> {
        let client = reqwest::Client::new()
            .get("https://tasks.googleapis.com/tasks/v1/users/@me/lists")
            .header("Authorization", format!("Bearer {}", self.access_token))
            .send()
            .await?;

        info!("{:?}", self.access_token);
        info!("{:?}", client);

        Ok(())
    }
}
