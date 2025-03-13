use std::collections::HashMap;

use anyhow::Result;
use reqwest::StatusCode;
use tracing::{error, info};

use crate::auth::google::GoogleOAuth;

const TASKS_API_ENDPOINT: &str = "https://tasks.googleapis.com/tasks/v1";

pub struct GoogleTasks {
    pub access_token: String,
    pub task_lists: Vec<String>,
    pub tasks: HashMap<String, Vec<String>>,
}

impl GoogleTasks {
    pub async fn new() -> Self {
        Self {
            access_token: GoogleOAuth::new().await.get_tokens().await.unwrap(),
            task_lists: Vec::new(),
            tasks: HashMap::new(),
        }
    }

    pub async fn get_tasks(self, _list: &str) -> Result<()> {
        unimplemented!()
    }

    pub async fn get_tasks_lists(self) -> Result<()> {
        let response = match reqwest::Client::new()
            .get(format!("{}{}", TASKS_API_ENDPOINT, "/users/@me/lists"))
            .header("Authorization", format!("Bearer {}", self.access_token))
            .send()
            .await
        {
            Err(e) => {
                error!("{}", e);
                serde_json::Value::default()
            }
            Ok(resp) => match resp.status() {
                StatusCode::OK => resp.json::<serde_json::Value>().await.unwrap(),
                StatusCode::UNAUTHORIZED => {
                    error!("Unauthenicated");
                    serde_json::Value::default()
                }
                _ => {
                    error!("{:?}", resp.json::<serde_json::Value>().await.ok());
                    serde_json::Value::default()
                }
            },
        };

        info!("{:?}", response);

        Ok(())
    }

    pub async fn add_tasks(self) -> Result<()> {
        unimplemented!()
    }
}
