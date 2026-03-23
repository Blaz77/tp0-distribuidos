from datetime import date, timedelta

from common.utils import Bet

class BetSerializer:
    @staticmethod
    def serialize(bet: Bet):
        raise NotImplementedError()
    
    @staticmethod
    def _deserialize_str(data: bytes, offset: int) -> tuple[str, int]:
        str_len = data[offset]
        offset += 1
        if str_len > 0:
            return data[offset:offset+str_len].decode("utf-8"), offset + str_len
        else:
            return "", offset
    
    @staticmethod
    def _deserialize_int(data: bytes, offset: int, signed=False) -> tuple[int, int]:
        INT_SIZE = 4
        return int.from_bytes(data[offset:offset+INT_SIZE], "big", signed=signed), offset + INT_SIZE
    
    @staticmethod
    def _deserialize_date(data: bytes, offset: int) -> tuple[date, int]:
        epoch = date(1970, 1, 1)
        days_since_epoch, offset = BetSerializer._deserialize_int(data, offset, signed=True)
        return epoch + timedelta(days=days_since_epoch), offset

    @staticmethod
    def deserialize(data: bytes) -> Bet:
        name, offset = BetSerializer._deserialize_str(data, 0)
        last_name, offset = BetSerializer._deserialize_str(data, offset)
        dni, offset = BetSerializer._deserialize_int(data, offset)
        birth_date, offset = BetSerializer._deserialize_date(data, offset)
        number, offset = BetSerializer._deserialize_int(data, offset)
        assert(offset == len(data))
        return Bet(str(1), name, last_name, str(dni), birth_date.isoformat(), str(number))