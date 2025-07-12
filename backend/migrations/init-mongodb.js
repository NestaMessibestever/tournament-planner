// Tournament Planner MongoDB Initialization Script
// Run this using: mongosh < init-mongodb.js

// Switch to tournament_planner database
use tournament_planner;

// Create collections with validation schemas
db.createCollection("activity_logs", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["tournament_id", "action", "timestamp"],
            properties: {
                tournament_id: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                user_id: {
                    bsonType: "string",
                    description: "user who performed the action"
                },
                action: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                entity_type: {
                    bsonType: "string",
                    description: "type of entity affected"
                },
                entity_id: {
                    bsonType: "string",
                    description: "ID of entity affected"
                },
                details: {
                    bsonType: "object",
                    description: "additional action details"
                },
                timestamp: {
                    bsonType: "date",
                    description: "must be a date and is required"
                }
            }
        }
    }
});

db.createCollection("match_updates", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["match_id", "tournament_id", "update_type", "timestamp"],
            properties: {
                match_id: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                tournament_id: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                update_type: {
                    enum: ["score", "status", "schedule", "venue", "referee"],
                    description: "type of update"
                },
                old_value: {
                    bsonType: ["object", "string", "number", "null"],
                    description: "previous value"
                },
                new_value: {
                    bsonType: ["object", "string", "number", "null"],
                    description: "new value"
                },
                updated_by: {
                    bsonType: "string",
                    description: "user who made the update"
                },
                timestamp: {
                    bsonType: "date",
                    description: "must be a date and is required"
                }
            }
        }
    }
});

db.createCollection("user_preferences", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["user_id"],
            properties: {
                user_id: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                theme: {
                    enum: ["light", "dark", "auto"],
                    description: "UI theme preference"
                },
                notifications: {
                    bsonType: "object",
                    properties: {
                        email: { bsonType: "bool" },
                        push: { bsonType: "bool" },
                        sms: { bsonType: "bool" }
                    }
                },
                language: {
                    bsonType: "string",
                    description: "preferred language code"
                },
                timezone: {
                    bsonType: "string",
                    description: "user's timezone"
                },
                dashboard_widgets: {
                    bsonType: "array",
                    description: "customized dashboard layout"
                }
            }
        }
    }
});

db.createCollection("websocket_connections", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["connection_id", "user_id", "connected_at"],
            properties: {
                connection_id: {
                    bsonType: "string",
                    description: "unique connection identifier"
                },
                user_id: {
                    bsonType: "string",
                    description: "connected user ID"
                },
                tournament_subscriptions: {
                    bsonType: "array",
                    items: {
                        bsonType: "string"
                    },
                    description: "tournaments this connection is subscribed to"
                },
                connected_at: {
                    bsonType: "date",
                    description: "connection timestamp"
                },
                last_ping: {
                    bsonType: "date",
                    description: "last heartbeat timestamp"
                }
            }
        }
    }
});

db.createCollection("tournament_analytics", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["tournament_id", "date"],
            properties: {
                tournament_id: {
                    bsonType: "string",
                    description: "tournament identifier"
                },
                date: {
                    bsonType: "date",
                    description: "analytics date"
                },
                views: {
                    bsonType: "int",
                    description: "page views count"
                },
                unique_visitors: {
                    bsonType: "int",
                    description: "unique visitors count"
                },
                registrations: {
                    bsonType: "int",
                    description: "new registrations count"
                },
                matches_completed: {
                    bsonType: "int",
                    description: "matches completed on this date"
                },
                average_match_duration: {
                    bsonType: "double",
                    description: "average match duration in minutes"
                }
            }
        }
    }
});

// Create indexes for better query performance
db.activity_logs.createIndex({ "tournament_id": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "user_id": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "entity_type": 1, "entity_id": 1 });

db.match_updates.createIndex({ "match_id": 1, "timestamp": -1 });
db.match_updates.createIndex({ "tournament_id": 1, "timestamp": -1 });
db.match_updates.createIndex({ "update_type": 1 });

db.user_preferences.createIndex({ "user_id": 1 }, { unique: true });

db.websocket_connections.createIndex({ "connection_id": 1 }, { unique: true });
db.websocket_connections.createIndex({ "user_id": 1 });
db.websocket_connections.createIndex({ "tournament_subscriptions": 1 });

db.tournament_analytics.createIndex({ "tournament_id": 1, "date": 1 }, { unique: true });

print("MongoDB initialization completed successfully!");
print("Collections created:");
print("- activity_logs");
print("- match_updates");
print("- user_preferences");
print("- websocket_connections");
print("- tournament_analytics");