import redis
from contextlib import contextmanager
from config.settings import settings

redis_client = redis.from_url(settings.REDIS_URL)

@contextmanager
def selenium_lock(timeout=60*10):
    """
    Context manager for acquiring and releasing a Redis lock.
    
    Args:
        - **lock_name** string: Name of the lock
        - **expire_time** int: Expiration time for the lock in seconds
    Yields:
        None
    """
    lock = redis_client.lock("selenium_lock", timeout=timeout)
    try:
        acquired = lock.acquire(blocking=True)
        if not acquired:
            raise Exception("Could not acquire the lock")
        
        yield
    finally:
        lock.release()