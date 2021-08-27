import pandas as pd

def merge(df1, df2):
    total = df1['No_of_Packets']+df2['No_of_Packets']
    result = {
        'second': range(601),
        'No_of_Packets': total[0:601]
        }
    res = pd.DataFrame(result,columns=['second','No_of_Packets'])
    res['No_of_Packets'] = res['No_of_Packets'].astype(int)
    return res

if __name__=='__main__':
    timeouts = [1, 2, 4]
    for t in timeouts:
        idle6633 = pd.read_csv('packet-in_report_6633_HARD_TIMEOUT_30_IDLE_TIMEOUT_'+str(t)+'.bench')
        idle6634 = pd.read_csv('packet-in_report_6634_HARD_TIMEOUT_30_IDLE_TIMEOUT_'+str(t)+'.bench')
        result = merge(idle6633, idle6634)
        result.to_csv('res_IDLE_'+str(t)+'.csv', index=False)
    
    tot = {'second':range(601)}
    for t in timeouts:
        df = pd.read_csv('res_IDLE_'+str(t)+'.csv')
        tot['IDLE_TIMEOUT='+str(t)] = df['No_of_Packets']
    
    total_df = pd.DataFrame(tot)
    total_df.to_csv('res_IDLE.csv', index=False)