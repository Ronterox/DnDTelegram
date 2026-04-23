import { Database } from "bun:sqlite";
import { User } from "./types";
export class DatabaseHelper {
    private db: Database;

    constructor(dbName: string = process.env.DB_PATH || "users.db") {
        this.db = new Database(dbName, { strict: true });
        this.init();
    }
    private init() {
        const createTableQuery = `
            CREATE TABLE IF NOT EXISTS users (
                email TEXT PRIMARY KEY,
                username TEXT NOT NULL,
                password TEXT NOT NULL

            );
        `;
        this.db.run(createTableQuery);
    }

    public getUser(email: string): User | null {
        const query = `
            SELECT * FROM users WHERE email = ?;
        `;
        const res = this.db.query(query).get(email) as User;

        return res ?? null;
    }

    public usertUser(user: User): void {
        const query = `
            INSERT INTO users (email, username, password)
            VALUES (?, ?, ?);
        `;
        this.db.query(query).run(user.email, user.username, user.password);
    }
}
