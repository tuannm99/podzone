mod root;
mod router;

pub use root::switch_root;
pub use router::{MainRoute, SettingsRoute, switch_main, switch_settings};
