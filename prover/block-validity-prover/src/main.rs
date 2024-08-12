extern crate actix;
extern crate actix_web;
extern crate config;

use std::env;

use actix_web::{middleware, web, App, HttpServer};
use actix_web_validator::{JsonConfig, PathConfig, QueryConfig};
use app::state::AppState;

use crate::app::error_handler::handle_error;

pub mod app;
pub mod proof;
pub mod server;

#[actix_rt::main]
async fn main() -> Result<(), std::io::Error> {
    let hostname: String = app::config::get("hostname");
    let port = env::var("PORT").expect("PORT must be set");
    let redis_url = env::var("REDIS_URL").expect("REDIS_URL must be set");
    let listen_address = format!("{}:{}", hostname, port);

    let redis = match redis::Client::open(redis_url.clone()) {
        Ok(client) => client,
        Err(e) => {
            eprintln!("Failed to create Redis client: {}", e);
            return Err(std::io::Error::new(
                std::io::ErrorKind::Other,
                "Failed to create Redis client",
            ));
        }
    };

    println!("Listening to requests at {}...", listen_address);
    HttpServer::new(move || {
        let state = AppState::new();

        App::new()
            .app_data(web::Data::new(redis.clone()))
            .app_data(web::Data::new(state))
            .app_data(PathConfig::default().error_handler(handle_error))
            .app_data(QueryConfig::default().error_handler(handle_error))
            .app_data(JsonConfig::default().error_handler(handle_error))
            .app_data(web::JsonConfig::default().limit(10_000_000)) // 10MB
            .configure(app::route::setup_routes)
            .wrap(middleware::Logger::default())
    })
    .bind(listen_address.clone())?
    .run()
    .await
}
