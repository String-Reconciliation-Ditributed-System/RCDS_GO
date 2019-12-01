# Recursive Content-Dependent Shingling (RCDS)
The RCDS [1] is a scalable string reconciliation protocol that is best used in distributed systems. This GO 
implementation relies on the implementation of [cpisync](https://github.com/trachten/cpisync) project. The RCDS 
breaks a file (string) into a set of data and relies on any set reconciliation primitives (CPI [2], Interactive 
CPI [3], and IBLT [4].) to reconcile the data and reconstruct the data back to a file. 

This implementation is a remote file synchronization utility. An existing C++ implementation is available, however, not 
well maintained on [forked cpisync](https://github.com/trachten/cpisync). 

## Reference:
If you use this work, please cite any relevant papers below.

[1] B. Song and A. Trachtenberg, "Scalable String Reconciliation by Recursive Content-Dependent Shingling"
      57th Annual Allerton Conference on Communication, Control, and Computing, 2019 
      [(Allerton)](https://proceedings.allerton.csl.illinois.edu/media/files/0073.pdf)  
 
[2] Y. Minsky, A. Trachtenberg, and R. Zippel,
     "Set Reconciliation with Nearly Optimal Communication Complexity",
     IEEE Transactions on Information Theory, 49:9.
     <http://ipsit.bu.edu/documents/ieee-it3-web.pdf>
     
[3] Y. Minsky and A. Trachtenberg,
     "Scalable set reconciliation"
     40th Annual Allerton Conference on Communication, Control, and Computing, 2002.
     <http://ipsit.bu.edu/documents/BUTR2002-01.pdf>
     

[4] Goodrich, Michael T., and Michael Mitzenmacher. "Invertible bloom lookup tables." 49th Annual Allerton Conference
 on Communication, Control, and Computing (Allerton), 2011. [arXiv](https://arxiv.org/abs/1101.2245)
