export function getFileUrl(path: string) {
  path = path.split("uploads")[1];
  return `http://localhost:4001/api/uploads${path}`;
}
