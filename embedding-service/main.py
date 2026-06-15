"""
GitSense Embedding Service
Python FastAPI sidecar for sentence-transformers inference
"""

import os
import logging
from typing import List, Optional

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("embedding-service")

# Configuration
MODEL_NAME = os.getenv("EMBEDDING_MODEL", "all-MiniLM-L6-v2")
MAX_BATCH_SIZE = int(os.getenv("MAX_BATCH_SIZE", "64"))
MAX_TEXT_LENGTH = int(os.getenv("MAX_TEXT_LENGTH", "8192"))

# Load model at startup (offline, from host volume cache)
logger.info(f"Loading model: {MODEL_NAME}")
model = SentenceTransformer(MODEL_NAME, local_files_only=True)
DIMENSIONS = model.get_sentence_embedding_dimension()
logger.info(f"Model loaded: {MODEL_NAME}, dimensions={DIMENSIONS}")

app = FastAPI(title="GitSense Embedding Service", version="1.0.0")


class EmbedRequest(BaseModel):
    text: str


class EmbedBatchRequest(BaseModel):
    texts: List[str]


class EmbedResponse(BaseModel):
    embedding: List[float]
    dimensions: int


class EmbedBatchResponse(BaseModel):
    embeddings: List[List[float]]
    dimensions: int


class HealthResponse(BaseModel):
    status: str
    model: str
    dimensions: int


@app.get("/health", response_model=HealthResponse)
async def health():
    return HealthResponse(
        status="ok",
        model=MODEL_NAME,
        dimensions=DIMENSIONS,
    )


@app.post("/embed", response_model=EmbedResponse)
async def embed(req: EmbedRequest):
    text = req.text[:MAX_TEXT_LENGTH]
    try:
        embedding = model.encode(text, normalize_embeddings=True)
        return EmbedResponse(
            embedding=embedding.tolist(),
            dimensions=DIMENSIONS,
        )
    except Exception as e:
        logger.error(f"Embedding error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/embed/batch", response_model=EmbedBatchResponse)
async def embed_batch(req: EmbedBatchRequest):
    if len(req.texts) > MAX_BATCH_SIZE:
        raise HTTPException(
            status_code=400,
            detail=f"Batch size {len(req.texts)} exceeds max {MAX_BATCH_SIZE}",
        )

    texts = [t[:MAX_TEXT_LENGTH] for t in req.texts]
    try:
        embeddings = model.encode(texts, normalize_embeddings=True, batch_size=32)
        return EmbedBatchResponse(
            embeddings=embeddings.tolist(),
            dimensions=DIMENSIONS,
        )
    except Exception as e:
        logger.error(f"Batch embedding error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8001"))
    uvicorn.run(app, host="0.0.0.0", port=port)
