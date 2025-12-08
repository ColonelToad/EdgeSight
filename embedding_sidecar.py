import os
os.environ['KMP_DUPLICATE_LIB_OK'] = 'TRUE'

from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

# Lazy-load models to avoid DLL issues at import time
_embed_model = None
_llm_model = None

def get_embed_model():
    global _embed_model
    if _embed_model is None:
        try:
            from sentence_transformers import SentenceTransformer
            _embed_model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')
        except Exception as e:
            print(f"Error loading embedding model: {e}")
            raise
    return _embed_model

def get_llm_model():
    global _llm_model
    if _llm_model is None:
        try:
            from transformers import AutoModelForCausalLM, AutoTokenizer
            print("Loading Qwen model (this may take a moment)...")
            _llm_model = {
                'model': AutoModelForCausalLM.from_pretrained(
                    "Qwen/Qwen2.5-1.5B-Instruct",
                    device_map="cpu",
                    torch_dtype="auto",
                ),
                'tokenizer': AutoTokenizer.from_pretrained("Qwen/Qwen2.5-1.5B-Instruct")
            }
            print("Qwen model loaded successfully")
        except Exception as e:
            print(f"Error loading LLM model: {e}")
            raise
    return _llm_model

app = FastAPI()

class EmbedRequest(BaseModel):
    text: str

class EmbedResponse(BaseModel):
    embedding: list[float]

class QueryRequest(BaseModel):
    system: str
    user: str
    max_tokens: int = 256

class QueryResponse(BaseModel):
    answer: str

@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest):
    model = get_embed_model()
    vec = model.encode([req.text])[0].tolist()
    return {"embedding": vec}

@app.post("/query", response_model=QueryResponse)
def query(req: QueryRequest):
    import torch
    llm_data = get_llm_model()
    model = llm_data['model']
    tokenizer = llm_data['tokenizer']
    
    # Build messages in Qwen chat format
    messages = [
        {"role": "system", "content": req.system},
        {"role": "user", "content": req.user}
    ]
    
    # Tokenize and generate
    text = tokenizer.apply_chat_template(
        messages,
        tokenize=False,
        add_generation_prompt=True
    )
    model_inputs = tokenizer([text], return_tensors="pt")
    
    # Generate response
    with torch.no_grad():
        generated_ids = model.generate(
            **model_inputs,
            max_new_tokens=req.max_tokens,
            temperature=0.2
        )
    
    # Decode response, removing the input prompt
    generated_ids = [output_ids[len(input_ids):] for input_ids, output_ids in zip(model_inputs.input_ids, generated_ids)]
    answer = tokenizer.batch_decode(generated_ids, skip_special_tokens=True)[0]
    
    return {"answer": answer.strip()}

@app.get("/health")
def health():
    return {"status": "ok"}

if __name__ == "__main__":
    print("Starting embedding + LLM sidecar on port 9000...")
    uvicorn.run(app, host="0.0.0.0", port=9000)
