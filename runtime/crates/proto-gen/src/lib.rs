// Include generated protobuf code.
// After running build.rs, generated files appear in src/gen/

pub mod platform {
    pub mod host_agent {
        pub mod v1 {
            include!("gen/platform.host_agent.v1.rs");
        }
    }
    pub mod identity {
        pub mod v1 {
            include!("gen/platform.identity.v1.rs");
        }
    }
    pub mod activity {
        pub mod v1 {
            include!("gen/platform.activity.v1.rs");
        }
    }
    pub mod guardrails {
        pub mod v1 {
            include!("gen/platform.guardrails.v1.rs");
        }
    }
    pub mod workspace {
        pub mod v1 {
            include!("gen/platform.workspace.v1.rs");
        }
    }
    pub mod human {
        pub mod v1 {
            include!("gen/platform.human.v1.rs");
        }
    }
    pub mod economics {
        pub mod v1 {
            include!("gen/platform.economics.v1.rs");
        }
    }
}
