import { getDB } from "@hasura/ndc-duckduckapi";

/**
 * Add a message to the database
 */
export async function addMessage(text: string): Promise<string> {
  const db = await getDB();

  await db.run(
    'INSERT INTO messages (text) VALUES (?)',
    [text]
  );

  return `Message added: ${text}`;
}

/**
 * Get all messages from the database
 */
export async function getMessages(): Promise<any[]> {
  const db = await getDB();

  const messages = await db.all('SELECT * FROM messages ORDER BY created_at DESC');
  return messages;
}

/**
 * Get message count
 */
export async function getMessageCount(): Promise<number> {
  const db = await getDB();

  const result = await db.get('SELECT COUNT(*) as count FROM messages');
  return result.count;
}
