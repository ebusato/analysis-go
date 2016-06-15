
TH1F* getHisto(TTree* t, TString name, TString cut) 
{
  TH1F* h = new TH1F("h"+name,"h"+name,100,200,800);
  t->Draw("PedestalSamples.Data >> h"+name, cut);
  return h;
}

void makePlot(TTree* tcpu, TTree* tlabv, TString name, TString cut)
{
  TH1F* hcpu = getHisto(tcpu, name+"cpu", cut);
  TH1F* hlabv = getHisto(tlabv, name+"labv", cut);
  hlabv->SetLineColor(kRed);
  hlabv->SetLineWidth(2);
  hcpu->SetLineWidth(2);
  hcpu->Draw();
  hlabv->Draw("same");
}

void plot() 
{
  gROOT->SetStyle("Plain");
  gStyle->SetOptStat(0);
  
  TFile* fcpu = new TFile("../calibConstants/pedestals_cpu.root");
  TTree* tcpu = (TTree*) fcpu->Get("tree");

  TFile* flabv = new TFile("../calibConstants/pedestals_labview.root");
  TTree* tlabv = (TTree*) flabv->Get("tree");

  TCanvas* c = new TCanvas("c","c",2000,1000);
  c->Divide(3,3);
  
  for(int i = 0; i < 9; ++i) {
    c->cd(i+1);
    makePlot(tcpu, tlabv, Form("000%d", i), Form("IDRS == 0 && IQuartet == 0 && IChannel == 0 && ICapacitor == %d", i));
  }

  c->SaveAs("output/pedestalsSummary.png");
}
