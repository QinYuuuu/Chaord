import random
import math
from sympy import mod_inverse
import numpy as np
import matplotlib.pyplot as plt
import decimal
import logging
import time
import argparse



def pow_normal(a, b, p):
    result = pow(a, b, p)
    half_p = (p - 1) // 2
    if result > half_p:
        result -= p
    return result

def mod_normal(a, p):
    result = a % p
    half_p = (p - 1) // 2
    if result > half_p:
        result -= p
    return result

def generate_coefficients(t, secret, p, normal = None):
    if normal is None:
        coeffs = [secret] + [random.randint(0, p - 1) for _ in range(t)]
    else:
        coeffs = [secret] + [random.randint(-(p - 1) / 2, (p - 1) / 2) for _ in range(t)]
    return coeffs

def polynomial(x, coeffs, p_mod, normal = None):
    result = 0
    for i, coeff in enumerate(coeffs):
            if normal is None:
                result = (int(result) + (int(coeff) * pow(x, i, p_mod) % p_mod)) % p_mod
            else:
                value = mod_normal(int(coeff) * pow_normal(x, i, p_mod), p_mod)
                result = mod_normal(int(result) + value, p_mod)
    return result

def share_secret(secret, n, t, p_mod, normal = None):
    coeffs = generate_coefficients(t, secret, p_mod, normal)
    shares = [(i, polynomial(i, coeffs, p_mod, normal)) for i in range(1, n + 1)]
    return shares

def reconstruct_secret(shares, t, p_mod):
    secret = 0
    for i in range(t):
        xi, yi = shares[i]
        li = 1
        if p_mod == 2:
            for j in range(t):
                if i != j:
                    xj, _ = shares[j]
                    li *= xj / (xj - xi)
            secret += yi * li 
            #print(secret,int(round(secret)))
        else:
            for j in range(t):
                if i != j:
                    xj, _ = shares[j]
                    li = li * (xj * mod_inverse(xj - xi, p_mod) % p_mod) % p_mod
            secret = (int(secret) + (int(yi) % p_mod) * (int(li) % p_mod) % p_mod) % p_mod
    return int(round(secret)) % p_mod


def mod_decimal_normal(a, b):
    a = a - (a // b) * b
    if a > b / 2:
        return a - b
    elif a < -b / 2:
        return a + b
    else:
        return a

def mod_inverse_normal(a, p):
    result = mod_inverse(a, p)
    half_p = (p - 1) // 2
    if result > half_p:
        result -= p
    return result

def reconstruct_secret_fmod(shares, t, p_mod):
    secret = 0
    for i in range(t):
        xi, yi = shares[i]
        li = 1
        for j in range(t):
            if i != j:
                xj, _ = shares[j]
                li = mod_decimal_normal(li * mod_decimal_normal(xj * mod_inverse_normal(xj - xi, p_mod), p_mod), p_mod)
        secret = mod_decimal_normal(decimal.Decimal(secret) + decimal.Decimal(yi) * decimal.Decimal(li), p_mod)
        # print(f'yi: {i, yi, li}')
        # print(f'secret: {i, secret}')
    return secret

def reconstruct_secret_fmod_normal(shares, t, p_mod, v):
    secret = 0
    for i in range(t):
        xi, yi = shares[i]
        li = 1
        for j in range(t):
            if i != j:
                xj, _ = shares[j]
                li = mod_decimal_normal(li * mod_decimal_normal(xj * mod_inverse_normal(xj - xi, p_mod), p_mod), p_mod)
        secret = mod_decimal_normal(decimal.Decimal(secret) + decimal.Decimal(yi) * decimal.Decimal(li), decimal.Decimal(p_mod) * v)
        # print(f'yi: {i, yi, li}')
        # print(f'secret: {i, secret}')
    return secret

def re_index(a):
    for i in range(len(a)):
        index, value = a[i]
        a[i] = (1 + i, value)
    return a

def add_shares(a, p):
    result = 0
    for _, value in a:
        result = (int(result) + int(value)) % p
    return result

def test_ODO(n, t, k, p, mean, var):
    # timer
    timer_node = 0
    timer_node_offline = 0

    batch = math.ceil(k / n)

    R_shares = np.empty((n, batch, n), dtype = object)
    R_square_shares = np.empty((n, batch, n), dtype = object)
    coin_pre = np.empty((n, batch), dtype=object)
    coin_pre_shares = np.empty((n, batch, n), dtype=object)
    coin = np.empty((n, batch), dtype=object)
    b_bias_shares = np.empty((n, batch, n), dtype=object)
    b_bias_add_R_shares = np.empty((n, batch, n), dtype=object)
    b_bias_add_R = np.empty((n, batch), dtype=object)
    b_bias_add_R_square_shares = np.empty((n, batch, n), dtype=object)
    MPC_shares = np.empty((n, batch, n), dtype=object)
    MPC = np.empty((n, batch), dtype=object)
    b_unbias_tmp_shares = np.empty((n, batch, n), dtype=object)
    b_unbias_shares = np.empty((n * batch, n), dtype=object)
    gaussian_non_normal_shares_Zp = np.empty(n, dtype=object)
    gaussian_shares_Zp = np.empty(n, dtype=object)

    # offline: 生成R的份额和R^2的份额
    time_tmp_begin = time.time()
    R = np.random.randint(0, 2, (n, batch))
    for node in range(n):
        for i in range(batch):
            R_shares[node, i, :] = share_secret(R[node, i], n, t, p, normal=True)
            R_square_shares[node, i, :] = share_secret(R[node, i], n, t, p, normal=True)
  
    for node in range(n):
        coin_pre[node] = np.random.randint(0, 2, batch)
        for i in range(batch):
            coin_pre_shares[node, i, :] = share_secret(coin_pre[node, i], n, t, p, normal=True)

    time_tmp_end = time.time()
    timer_node_offline += (time_tmp_end - time_tmp_begin) / n * 1000


    time_tmp_begin = time.time()    

    for node in range(n):
        for i in range(batch):
            coin[node, i] = reconstruct_secret_fmod_normal(coin_pre_shares[node, i, :], t + 1, p, 1)

    time_tmp_end = time.time()
    timer_node_offline += (time_tmp_end - time_tmp_begin) * 1000
    
    

    # step 1: 节点生成k/n个比特的份额
    time_tmp_begin = time.time()

    for node in range(n):
        b_bias = np.random.randint(0, 2, batch)
        for i in range(batch):
            b_bias_shares[node, i, :] = share_secret(b_bias[i], n, t, p, normal=True)
    
    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000

    # step 2: 节点验证b^2=b
    time_tmp_begin = time.time()

    for node in range(n):
        for i in range(batch):
            for j in range(n):
                b_bias_add_R_shares[node, i, j] = (j + 1, b_bias_shares[node, i, j][1] + R_shares[node, i, j][1])

    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000


    time_tmp_begin = time.time()
    for i in range(batch):
        for node in range(n):
            b_bias_add_R[node, i] = reconstruct_secret_fmod_normal(b_bias_add_R_shares[node, i, :], t + 1, p, 1)
    
    for i in range(batch):
        for node in range(n):
            value = b_bias_add_R[node, i] * b_bias_add_R[node, i]
            b_bias_add_R_square_shares[node, i, :] = share_secret(value, n, t, p, normal=True)


    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) * 1000



    time_tmp_begin = time.time()
    for node in range(n):
        for i in range(batch):
            for j in range(n):
                value = b_bias_add_R_square_shares[node, i, j][1] - 2 * R_shares[node, i, j][1] * b_bias_add_R[node, i] - b_bias_shares[node, i, j][1] + R_square_shares[node, i, j][1]
                MPC_shares[node, i, j] = (j + 1, value)
    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000

    time_tmp_begin = time.time()
    for i in range(batch):
        for node in range(n):
            MPC[node, i] = reconstruct_secret_fmod_normal(MPC_shares[node, i, :], t + 1, p, 1)

    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) * 1000

    for i in range(batch):
        for node in range(n):
            if MPC[node, i] != 0:
                logging.error('ODO - MPC error')

    # step 3: 翻转
    time_tmp_begin = time.time()
    for node in range(n):
        for i in range(batch):
            for j in range(n):
                if coin[node, i] == 0:
                    b_unbias_tmp_shares[node, i, j] = b_bias_shares[node, i, j]
                else:
                    index, value = b_bias_shares[node, i, j]
                    value = mod_normal(1 - value, p)
                    b_unbias_tmp_shares[node, i, j] = (index, value)

    for node in range(n):
        b_unbias_shares[:, node] = b_unbias_tmp_shares[:, :, node].flatten()
    
    # 求和
    for node in range(n):
        gaussian_non_normal_shares_Zp[node] = (node + 1, add_shares(b_unbias_shares[:, node], p))
    
    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000

    # 标准化
    time_tmp_begin = time.time()
    
    v1 = decimal.Decimal(math.sqrt(4 * var / k))
    mean = decimal.Decimal(mean)
    v2 = - v1 * k / 2 + mean
    v2_shares = share_secret(k / 2, n, t, p)
    for node in range(n):
        _, gaussian_value = gaussian_non_normal_shares_Zp[node]
        _, v2_value = v2_shares[node]
        value =  v1 * mod_decimal_normal(decimal.Decimal(gaussian_value) - v2_value, decimal.Decimal(p))
        gaussian_shares_Zp[node] = (node + 1, value)
    
    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000

    gaussian_normal = reconstruct_secret_fmod_normal(gaussian_shares_Zp, t + 1, p, v1)
    
    return gaussian_normal, timer_node, timer_node_offline


def test_trading_online_Zp_fmod(n, t, k, p, mean, var):
    # 测试data trading中三个参与方的运行时间

    # print(f'============================= online begin ================================')

    # timer
    timer_node = 0
    timer_owner = 0
    timer_consumer = 0
    timer_node_offline = 0

    # sets consumer - 1, owner - 0
    gaussian_non_normal_shares_Zp = np.empty(n, dtype=object)
    gaussian_shares_Zp = np.empty(n, dtype=object)

    batch = math.ceil(k / n)
    R_shares = np.empty((n), dtype = object)
    R_square_shares = np.empty((n), dtype = object)
    coin_pre = np.empty((n, batch), dtype=object)
    coin_pre_shares = np.empty((n, batch, n), dtype=object)
    coin = np.empty((n, batch), dtype=object)
    bc_shares = np.empty((n), dtype=object)
    bo_shares = np.empty((n), dtype=object)
    bc_add_bo_shares = np.empty((n), dtype=object)
    b_unbias_shares = np.empty((k, n), dtype=object)

    bo_add_R_shares = np.empty((n), dtype=object)
    bo_add_R = np.empty(1, dtype=object)
    bo_add_R_square_shares = np.empty((n), dtype=object)
    MPC_shares = np.empty((n), dtype=object)
    MPC = np.empty((1), dtype=object)
    
    # offline
    time_tmp_begin = time.time()    
    for node in range(n):
        coin_pre[node] = np.random.randint(0, 2, batch)
        for i in range(batch):
            coin_pre_shares[node, i, :] = share_secret(coin_pre[node, i], n, t, p, normal=True)

    R = np.random.randint(0, 2)
    R_shares = share_secret(R, n, t, p, normal=True)
    R_square_shares = share_secret((R * R) % 2, n, t, p, normal=True)

    time_tmp_end = time.time()
    timer_node_offline += (time_tmp_end - time_tmp_begin) / n * 1000

    time_tmp_begin = time.time()    

    for node in range(n):
        for i in range(batch):
            coin[node, i] = reconstruct_secret_fmod_normal(coin_pre_shares[node, i, :], t + 1, p, 1)
    coin = coin.flatten()

    time_tmp_end = time.time()
    timer_node_offline += (time_tmp_end - time_tmp_begin) * 1000

    # step 1: consumer选择1个随机数𝑏c∈ℤ_2，ℤ_2上秘密分享给committee node

    # consumer
    time_tmp_begin = time.time()

    bc = np.random.randint(0, 2)
    bc_shares = share_secret(bc, n, t, 2)
    
    time_tmp_end = time.time()
    timer_consumer += (time_tmp_end - time_tmp_begin) * 1000

    #print(f'============================= step 1 done ================================')

    # step 2: owner选择1个随机数𝑏o∈ℤ_2，ℤ_p上秘密分享给committee node

    # owner
    time_tmp_begin = time.time()

    bo = np.random.randint(0, 2)
    bo_shares = share_secret(bo, n, t, p, normal=True)

    time_tmp_end = time.time()
    timer_owner += (time_tmp_end - time_tmp_begin) * 1000

    # node verify b^2-b
    time_tmp_begin = time.time()

    for node in range(n):
        bo_add_R_shares[node] = (node + 1, bo_shares[node][1] + R_shares[node][1])

    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000


    time_tmp_begin = time.time()
    bo_add_R = reconstruct_secret_fmod_normal(bo_add_R_shares, t + 1, p, 1)

    value = bo_add_R * bo_add_R
    bo_add_R_square_shares = share_secret(value, n, t, p, normal=True)


    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) * 1000

    bo_add_R_square = reconstruct_secret_fmod_normal(bo_add_R_square_shares, t + 1, p, 1)

    if bo_add_R != bo + R :
        logging.error('Chaord - bo_add_R error')

    if bo_add_R_square != (bo + R % 2) ** 2 :
        logging.error(f'Chaord - bo_add_R_square error, {bo_add_R_square}, {(bo + R % 2) ** 2}')


    time_tmp_begin = time.time()
    for node in range(n):
        value = (bo_add_R_square_shares[node][1] - 2 * R_shares[node][1] * bo_add_R - bo_shares[node][1] + R_square_shares[node][1]) % p
        MPC_shares[node] = (node + 1, value)
    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000

    time_tmp_begin = time.time()
    MPC = reconstruct_secret_fmod_normal(MPC_shares, t + 1, p, 1)

    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) * 1000
    
    if MPC != 0 :
        logging.error(f'Chaord - MPC error, {MPC}')

    
    #print(f'============================= step 2 done ================================')


    # step 3: 重构consumer bit
    # node

    time_tmp_begin = time.time()
    bc_rec = reconstruct_secret(bc_shares, t + 1, 2)

    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) * 1000

    if bc != bc_rec :
        logging.error('Chaord - bc reconstruct error')


    #print(f'============================= step 3 done ================================')

    # step 4: 得到1个无偏比特份额

    # node
    time_tmp_begin = time.time()

    

    for node in range(n):
        if bc_rec == 0:
            bc_add_bo_shares[node] = bo_shares[node]
        elif bc_rec == 1:
            index, value = bo_shares[node]
            value = mod_normal(1 - value, p)
            bc_add_bo_shares[node] = (index, value)
        #print(bo_shares, bc_rec, bc_add_bo_shares)

    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000


    bc_add_bo = reconstruct_secret_fmod_normal(bc_add_bo_shares, t + 1, p, True)
    if bc_add_bo != (bc + bo) % 2:
        logging.error(f'Chaord - bc_add_bo bit reconstruct error, {bc_add_bo}, {bc + bo % 2}')

    # step 5：得到k个无偏比特份额
    time_tmp_begin = time.time()
    for i in range(k):
        for node in range(n):
            if coin[i] == 0:
                b_unbias_shares[i, node] = bc_add_bo_shares[node]
            else:
                index, value = bc_add_bo_shares[node]
                value = mod_normal(1 - value, p)
                b_unbias_shares[i, node] = (index, value)
    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000




    # step 6: 求和得到高斯份额

    # node
    time_tmp_begin = time.time()

    for node in range(n):
        gaussian_non_normal_shares_Zp[node] = (node + 1, add_shares(b_unbias_shares[:, node], p))

    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000

    
    # # step 9: 标准化
    # for node in range(n):
    #     gaussian_shares_Zp[node] = (node + 1, gaussian_non_normal_shares_Zp[node][1])
    # gaussian_non_normal = reconstruct_secret(gaussian_shares_Zp, t + 1, p)

    # step 9: 标准化
    time_tmp_begin = time.time()
    
    v1 = decimal.Decimal(math.sqrt(4 * var / k))
    mean = decimal.Decimal(mean)
    v2 = - v1 * k / 2 + mean
    v2_shares = share_secret(k / 2, n, t, p)
    for node in range(n):
        _, gaussian_value = gaussian_non_normal_shares_Zp[node]
        _, v2_value = v2_shares[node]
        value =  v1 * mod_decimal_normal(decimal.Decimal(gaussian_value) - v2_value, decimal.Decimal(p))
        gaussian_shares_Zp[node] = (node + 1, value)
    
    time_tmp_end = time.time()
    timer_node += (time_tmp_end - time_tmp_begin) / n * 1000


    #print(f'============================= step 8 done ================================')
            
    # step : reconstruct test
    
    gaussian_non_normal = reconstruct_secret_fmod_normal(gaussian_shares_Zp, t + 1, p, v1)
    
    # if gaussian_non_normal > p / 2: 
    #      gaussian_non_normal = gaussian_non_normal - p
    #print(gaussian_non_normal)
    
    return gaussian_non_normal, timer_consumer, timer_node, timer_owner, timer_node_offline

def parse_args():
    parser = argparse.ArgumentParser(description="Test ODO and Data Trading")
    parser.add_argument("--expNum", type=int, default=3072, help="Number of experiments")
    parser.add_argument("--p", type=int, default=104876876361812059901865976035638639526659825757836510784054252380955342739513, help="Modulus value")
    parser.add_argument("--n", type=int, default=31, help="Number of nodes")
    parser.add_argument("--t", type=int, default=10, help="Threshold value")
    parser.add_argument("--k", type=int, default=500, help="Distribution parameter")
    parser.add_argument("--mean", type=float, default=0.0, help="Mean value")
    parser.add_argument("--var", type=float, default=4.0, help="Variance value")
    return parser.parse_args()


if __name__ == "__main__":
    # Parse arguments
    args = parse_args()

    # Retrieve parameters from the parsed arguments
    expNum = args.expNum
    p = args.p
    n = args.n
    t = args.t
    k = args.k
    mean = args.mean
    var = args.var

    # Configure logging
    logging.basicConfig(filename=f'output.log', level=logging.INFO, 
                    format='%(asctime)s - %(levelname)s - %(message)s')
    logging.info('=====================================================')
    logging.info(f'setting: node number-{n}, threshold-{t}, distribution parameter-{k}, data scale-{expNum}, var-{var}, mean-{mean}, order-256bit')

    # decimal
    decimal.getcontext().prec = 200

    # log
    logging.basicConfig(level=logging.ERROR, format='%(levelname)s: %(message)s')

    # ODO
    # data trading
    gaussian_non_normal_ODO = np.empty(expNum)
    timer_consumer_ODO = np.empty(expNum)
    timer_node_ODO = np.empty(expNum)
    timer_owner_ODO = np.empty(expNum)
    timer_node_offline_ODO = np.empty(expNum)

    for i in range(expNum):
        gaussian_non_normal_ODO[i], timer_node_ODO[i], timer_node_offline_ODO[i]= test_ODO(n, t, k, p, mean, var)
        if i > 1 and i % 100 == 0:
            print(f'=============== {i + 1}-th ODO =================')
            print(f'mean: {np.mean(gaussian_non_normal_ODO[0:i]):.4f}')
            print(f'variance: {np.var(gaussian_non_normal_ODO[0:i]):.4f}')

    # data trading
    gaussian_non_normal = np.empty(expNum)
    timer_consumer = np.empty(expNum)
    timer_node = np.empty(expNum)
    timer_owner = np.empty(expNum)
    timer_node_offline = np.empty(expNum)

    for i in range(expNum):
        gaussian_non_normal[i], timer_consumer[i], timer_node[i], timer_owner[i], timer_node_offline[i]= test_trading_online_Zp_fmod(n, t, k, p, mean, var)
        if i > 1 and i % 100 == 0:
            print(f'=============== {i + 1}-th trading =================')
            print(f'mean: {np.mean(gaussian_non_normal[0:i]):.4f}')
            print(f'variance: {np.var(gaussian_non_normal[0:i]):.4f}')



    # Logging the information instead of printing
    logging.info('======================== ODO ========================')
    logging.info(f'mean_standard: {mean}')
    logging.info(f'variance_standard: {var}')

    logging.info(f'mean_gen_normalize: {np.mean(gaussian_non_normal_ODO):.4f}')
    logging.info(f'variance_gen_normalize: {np.var(gaussian_non_normal_ODO):.4f}')
    logging.info(f'parameters: {n} nodes, threshold {t}, distribution parameter {k}')

    logging.info('-------- average runtime -------')
    logging.info(f'node-online: {np.mean(timer_node_ODO)} ms')
    logging.info(f'node-offline: {np.mean(timer_node_offline_ODO)} ms')

    logging.info(f'-------- total runtime for {expNum} samples --------')
    logging.info(f'node-online: {np.sum(timer_node_ODO)} ms')
    logging.info(f'node-offline: {np.sum(timer_node_offline_ODO)} ms')

    logging.info('======================= data trading ====================')
    logging.info(f'mean_standard: {mean}')
    logging.info(f'variance_standard: {var}')

    logging.info(f'mean_gen_normalize: {np.mean(gaussian_non_normal):.4f}')
    logging.info(f'variance_gen_normalize: {np.var(gaussian_non_normal):.4f}')

    logging.info(f'parameters: {n} nodes, threshold {t}, distribution parameter {k}')

    logging.info('-------- average runtime --------')
    logging.info(f'consumer: {np.mean(timer_consumer)} ms')
    logging.info(f'owner: {np.mean(timer_owner)} ms')
    logging.info(f'node-online: {np.mean(timer_node)} ms')
    logging.info(f'node-offline: {np.mean(timer_node_offline)} ms')

    logging.info(f'-------- total runtime for {expNum} samples --------')
    logging.info(f'consumer: {np.sum(timer_consumer)} ms')
    logging.info(f'owner: {np.sum(timer_owner)} ms')
    logging.info(f'node-online: {np.sum(timer_node)} ms')
    logging.info(f'node-offline: {np.sum(timer_node_offline)} ms')
    



    
