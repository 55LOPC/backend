import base58
import axolotl_curve25519 as curve
import sys
import os
import time
import struct

# 
count = -7

recipient = sys.argv[count]
pubKey = sys.argv[count + 1]
privKey = sys.argv[count + 2] #sys.argv[-2]
amount = int(sys.argv[count + 3])
txfee = int(sys.argv[count + 4])
timestamp = int(sys.argv[count + 5])
attachment = sys.argv[count + 6]

lenAtt = len(attachment)

sData = '\4' + \
base58.b58decode(pubKey) + \
'\0\0' + \
struct.pack(">Q", timestamp) + \
struct.pack(">Q", amount) + \
struct.pack(">Q", txfee) + base58.b58decode(recipient) + struct.pack(">H", lenAtt) + attachment

random64 = os.urandom(64)
signature = base58.b58encode(curve.calculateSignature(random64, base58.b58decode(privKey), sData))
print(signature)
